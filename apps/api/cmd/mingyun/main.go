package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
	_ "time/tzdata" // 内嵌时区库，容器里没有 tzdata 也能解析 Asia/Shanghai
	"unicode"

	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/ai"
	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/entropy"
	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/httpapi"
	"github.com/huayuehua15-netizen/mingyun/apps/api/internal/store"
	"github.com/huayuehua15-netizen/mingyun/paipan/calendar"
	"github.com/huayuehua15-netizen/mingyun/paipan/daliuren"
	"github.com/huayuehua15-netizen/mingyun/paipan/liuyao"
	"github.com/huayuehua15-netizen/mingyun/paipan/schema"
	"github.com/huayuehua15-netizen/mingyun/paipan/ziwei"
)

// version 由构建时注入，默认开发版本。
var version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "version":
		fmt.Printf("mingyun %s\n", version)
	case "calendar":
		if err := runCalendar(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "错误:", err)
			os.Exit(1)
		}
	case "paipan":
		if err := runPaipan(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "错误:", err)
			os.Exit(1)
		}
	case "serve":
		if err := runServe(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "错误:", err)
			os.Exit(1)
		}
	case "migrate":
		fmt.Fprintln(os.Stderr, "数据库迁移将在 M2 接入（goose + embed）。当前无数据库依赖。")
		os.Exit(2)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "未知子命令: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `mingyun — 命云排盘引擎

用法:
  mingyun version                    输出版本
  mingyun calendar [参数]            计算历法上下文（干支、节气、月将、旬空）
  mingyun paipan [参数]              三术排盘，输出盘面 JSON
  mingyun serve [参数]               启动 HTTP 服务

历法参数（calendar / paipan 通用）:
  -date string        公历日期 YYYY-MM-DD，默认今天
  -time string        本地时间 HH:MM，默认 00:00
  -tz string          IANA 时区，默认 Asia/Shanghai
  -lon float         东经度数，默认 120（东八区中央经线）
  -true-solar bool   启用真太阳时，默认 true
  -boundary string    日界规则 zi_chu | zi_zheng，默认 zi_chu

paipan 参数:
  -method string      术数：liuyao | daliuren | ziwei
  -question string    所问之事
  -lines string       六爻，自下而上，逗号分隔：少阳,少阴,老阳,老阴
  -coins string       六组铜钱，逗号分隔，每组三位，1 为背：111,110,100,000,111,100
  -summary            只打印一行摘要，不输出完整 JSON

紫微参数:
  -gender string      性别：male / female
  -calendar string    历法：solar / lunar，默认 solar
  -birth-date string  出生日期 YYYY-MM-DD
  -birth-time string  出生时刻 HH:MM
  -leap-month        农历闰月
  -leap-policy string 闰月策略 as_this_month | as_next_month
`)
}

// calendarOptions 历法子命令共用的参数。
type calendarOptions struct {
	date      string
	clock     string
	tz        string
	lon       float64
	trueSolar bool
	boundary  string
}

func (o *calendarOptions) register(fs *flag.FlagSet) {
	fs.StringVar(&o.date, "date", "", "公历日期 YYYY-MM-DD")
	fs.StringVar(&o.clock, "time", "00:00", "本地时间 HH:MM")
	fs.StringVar(&o.tz, "tz", "Asia/Shanghai", "IANA 时区")
	fs.Float64Var(&o.lon, "lon", 120, "东经度数")
	fs.BoolVar(&o.trueSolar, "true-solar", true, "启用真太阳时")
	fs.StringVar(&o.boundary, "boundary", "zi_chu", "日界规则 zi_chu | zi_zheng")
}

func (o *calendarOptions) build() (calendar.Context, error) {
	loc, err := time.LoadLocation(o.tz)
	if err != nil {
		return calendar.Context{}, fmt.Errorf("加载时区 %q: %w", o.tz, err)
	}
	if o.date == "" {
		o.date = time.Now().In(loc).Format("2006-01-02")
	}
	at, err := time.ParseInLocation("2006-01-02 15:04", o.date+" "+o.clock, loc)
	if err != nil {
		return calendar.Context{}, fmt.Errorf("解析日期时间: %w", err)
	}
	db, err := calendar.ParseDayBoundary(o.boundary)
	if err != nil {
		return calendar.Context{}, err
	}
	cal, err := calendar.Build(calendar.Request{
		At:               at,
		Longitude:        o.lon,
		HasLongitude:     true,
		UseTrueSolarTime: o.trueSolar,
		DayBoundary:      db,
	})
	if err != nil {
		return calendar.Context{}, err
	}
	return cal, nil
}

func runCalendar(args []string) error {
	fs := flag.NewFlagSet("calendar", flag.ContinueOnError)
	var opts calendarOptions
	opts.register(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	cal, err := opts.build()
	if err != nil {
		return err
	}
	return encodeJSON(cal)
}

func runPaipan(args []string) error {
	fs := flag.NewFlagSet("paipan", flag.ContinueOnError)
	var opts calendarOptions
	opts.register(fs)
	method := fs.String("method", "liuyao", "术数：liuyao | daliuren | ziwei")
	question := fs.String("question", "", "所问之事")
	lines := fs.String("lines", "", "六爻：少阳,少阴,老阳,老阴")
	coins := fs.String("coins", "", "六组铜钱：111,110,100,000,111,100")
	summary := fs.Bool("summary", false, "只打印摘要")
	gender := fs.String("gender", "", "紫微性别：male / female")
	calendarType := fs.String("calendar", "solar", "紫微历法：solar / lunar")
	birthDate := fs.String("birth-date", "", "紫微出生日期 YYYY-MM-DD")
	birthTime := fs.String("birth-time", "", "紫微出生时刻 HH:MM")
	leapMonth := fs.Bool("leap-month", false, "紫微农历闰月")
	leapPolicy := fs.String("leap-policy", "as_this_month", "紫微闰月策略 as_this_month | as_next_month")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var chart schema.Chart
	var err error
	switch *method {
	case "liuyao":
		chart, err = castCLILiuYao(opts, *question, *lines, *coins)
	case "daliuren":
		cal, calErr := opts.build()
		if calErr != nil {
			return calErr
		}
		chart, err = daliuren.Cast(daliuren.Input{Question: *question}, cal)
	case "ziwei":
		loc, locErr := time.LoadLocation(opts.tz)
		if locErr != nil {
			return fmt.Errorf("加载时区 %q: %w", opts.tz, locErr)
		}
		birth, birthErr := ziwei.ParseBirth(ziwei.Input{
			Gender:          *gender,
			CalendarType:    *calendarType,
			BirthDate:       *birthDate,
			BirthTime:       *birthTime,
			IsLeapMonth:     *leapMonth,
			LeapMonthPolicy: *leapPolicy,
		}, ziwei.BirthOptions{
			Location:         loc,
			Longitude:        opts.lon,
			HasLongitude:     true,
			UseTrueSolarTime: opts.trueSolar,
		})
		if birthErr != nil {
			return birthErr
		}
		db, dbErr := calendar.ParseDayBoundary(opts.boundary)
		if dbErr != nil {
			return dbErr
		}
		cal, calErr := calendar.Build(calendar.Request{
			At:          birth.SolarTime,
			DayBoundary: db,
		})
		if calErr != nil {
			return calErr
		}
		chart, err = ziwei.Cast(ziwei.Input{
			Question:        *question,
			Gender:          *gender,
			CalendarType:    *calendarType,
			BirthDate:       *birthDate,
			BirthTime:       *birthTime,
			IsLeapMonth:     *leapMonth,
			LeapMonthPolicy: *leapPolicy,
		}, birth, cal)
	default:
		return fmt.Errorf("暂只支持 liuyao、daliuren、ziwei，收到 %q", *method)
	}
	if err != nil {
		return err
	}

	if *summary {
		switch p := chart.Payload.(type) {
		case liuyao.Payload:
			fmt.Println(p.Summary())
		case daliuren.Payload:
			fmt.Println(p.Summary())
		case ziwei.Payload:
			fmt.Println(p.Summary())
		}
		return nil
	}
	return encodeJSON(chart)
}

func castCLILiuYao(opts calendarOptions, question, lines, coins string) (schema.Chart, error) {
	in := liuyao.Input{Question: question}
	switch {
	case lines != "":
		in.Mode = liuyao.ModeManual
		for _, s := range splitList(lines) {
			k, err := liuyao.ParseLineKind(s)
			if err != nil {
				return schema.Chart{}, err
			}
			in.Lines = append(in.Lines, k)
		}
	case coins != "":
		in.Mode = liuyao.ModeCoin
		groups := splitList(coins)
		if len(groups) != 6 {
			return schema.Chart{}, fmt.Errorf("铜钱需 6 组，实为 %d", len(groups))
		}
		for i, g := range groups {
			if len([]rune(g)) != 3 {
				return schema.Chart{}, fmt.Errorf("第 %d 组铜钱需三位 0/1，实为 %q", i+1, g)
			}
			backs := 0
			for _, r := range []rune(g) {
				switch r {
				case '1':
					backs++
				case '0':
				default:
					return schema.Chart{}, fmt.Errorf("第 %d 组铜钱出现非法字符 %q", i+1, string(r))
				}
			}
			in.Backs[i] = backs
		}
	default:
		return schema.Chart{}, fmt.Errorf("请用 -lines 或 -coins 指定卦")
	}

	cal, err := opts.build()
	if err != nil {
		return schema.Chart{}, err
	}
	return liuyao.Cast(in, cal)
}

// splitList 按逗号或空白切分，兼容中文逗号与 PowerShell 的数组展开。
func splitList(s string) []string {
	s = strings.NewReplacer("，", ",", "、", ",").Replace(s)
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
}

func encodeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", defaultAddr(), "监听地址，如 :8080")
	if err := fs.Parse(args); err != nil {
		return err
	}

	secret := []byte(os.Getenv("APP_SECRET"))
	if len(secret) < 16 {
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			return fmt.Errorf("生成临时密钥: %w", err)
		}
		secret = buf
		fmt.Fprintln(os.Stderr, "警告：未设置 APP_SECRET（至少 16 字节），本次使用临时密钥；重启后 entropy_proof 将无法复核")
	}
	es, err := entropy.NewSource(secret)
	if err != nil {
		return err
	}

	aiCfg := ai.LoadConfig()
	var aiClient *ai.Client
	if aiCfg.Enabled() {
		aiClient = ai.NewClient(aiCfg)
		fmt.Printf("AI 解读已启用: provider=%s model=%s\n", aiCfg.Provider(), aiCfg.Model)
	} else {
		fmt.Println("AI 解读未配置：使用离线盘面摘要；设置 MINGYUN_AI_BASE_URL 与 MINGYUN_AI_MODEL 后启用")
	}

	srv := &http.Server{
		Addr:              *addr,
		Handler:           httpapi.New(version, store.NewMemory(), es, aiClient),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("mingyun %s 正在监听 %s\n", version, *addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		fmt.Println("收到退出信号，正在关闭…")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func defaultAddr() string {
	if port := os.Getenv("APP_PORT"); port != "" {
		return ":" + port
	}
	return ":8080"
}
