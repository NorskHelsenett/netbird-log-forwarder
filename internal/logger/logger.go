package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Splunk loggers
	SplunkTraffic *zap.SugaredLogger
	SplunkAudit   *zap.SugaredLogger

	// App logger (console + file)
	Log *zap.SugaredLogger
)

type SplunkHECWriter struct {
	URL        string
	Token      string
	Index      string
	Source     string
	SourceType string
	Host       string
	Client     *resty.Client
	PrintBody  bool // set true to debug payloads
}

func NewSplunkHECWriter(url, token, index, source, sourcetype, host string, timeout time.Duration) *SplunkHECWriter {
	return &SplunkHECWriter{
		URL:        url,
		Token:      token,
		Index:      index,
		Source:     source,
		SourceType: sourcetype,
		Host:       host,
		Client:     resty.New().SetTimeout(timeout),
		PrintBody:  false,
	}
}

func (w *SplunkHECWriter) Write(p []byte) (n int, err error) {
	var logTimestamp int64
	var logEntry map[string]interface{}

	// Zap event inn (JSON)
	if err := json.Unmarshal(p, &logEntry); err != nil {
		logTimestamp = time.Now().Unix()
	} else {
		// Hent timestamp fra "time" (se encoder under)
		if tsStr, ok := logEntry["time"].(string); ok {
			// ISO8601 fra zapcore.ISO8601TimeEncoder -> 2006-01-02T15:04:05.000Z0700
			t, parseErr := time.Parse(time.RFC3339Nano, tsStr)
			if parseErr != nil {
				// fallback – enkelte ISO8601-varianter
				if t2, e2 := time.Parse("2006-01-02T15:04:05.000Z0700", tsStr); e2 == nil {
					logTimestamp = t2.Unix()
				} else {
					logTimestamp = time.Now().Unix()
				}
			} else {
				logTimestamp = t.Unix()
			}
		} else {
			logTimestamp = time.Now().Unix()
		}

		// Rydd vekk felt vi ikke vil sende i "event"
		delete(logEntry, "time")
		delete(logEntry, "caller")
		delete(logEntry, "meta")
	}

	payload := map[string]interface{}{
		"event":      logEntry,
		"time":       logTimestamp,
		"host":       w.Host,
		"source":     w.Source,
		"sourcetype": w.SourceType,
		"index":      w.Index,
	}

	if w.PrintBody {
		pretty, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println("Sending to Splunk:", string(pretty))
	}

	resp, reqErr := w.Client.R().
		SetHeader("Authorization", "Splunk "+w.Token).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(w.URL + "/services/collector/event")

	// Ikke blokker appen ved feil
	if reqErr != nil {
		return len(p), nil
	}
	if resp.StatusCode() != 200 {
		fmt.Println("Failed to send to Splunk:", resp.Status())
	}
	return len(p), nil
}

func InitLogger(logDir string) error {
	// --- Encodere ---
	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:       "time", // VIKTIG: matcher writer
		LevelKey:      "level",
		MessageKey:    "message",
		CallerKey:     "caller",
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
		LineEnding:    zapcore.DefaultLineEnding,
		StacktraceKey: "",
		NameKey:       "",
		FunctionKey:   "",
	})

	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:      "time",
		LevelKey:     "level",
		MessageKey:   "message",
		CallerKey:    "caller",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.CapitalColorLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	})

	// --- APP: fil + konsoll ---
	lumberJack := &lumberjack.Logger{
		Filename:   logDir + "/app.log",
		MaxSize:    1000,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   false,
	}
	fileCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(lumberJack), zapcore.InfoLevel)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
	appCore := zapcore.NewTee(fileCore, consoleCore)
	Log = zap.New(appCore, zap.AddCaller()).Sugar()

	host := viper.GetString("splunk_host")
	if host == "" {
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "unknown-host"
		}
		host = hostname
	}

	// --- SPLUNK TRAFFIC ---
	splunkURL := viper.GetString("splunk.url")
	splunkToken := viper.GetString("splunk.traffic_token")
	splunkSource := viper.GetString("splunk.traffic_source")
	if splunkSource == "" {
		splunkSource = "netbird"
	}

	trafficIndex := viper.GetString("splunk.traffic_index")
	trafficSourceType := viper.GetString("splunk.traffic_source_type")
	if trafficSourceType == "" {
		trafficSourceType = "netbird:traffic"
	}

	if splunkURL != "" && splunkToken != "" && trafficIndex != "" {
		trafficWriter := NewSplunkHECWriter(splunkURL, splunkToken, trafficIndex, splunkSource, trafficSourceType, host, 5*time.Second)
		trafficCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(trafficWriter), zapcore.InfoLevel)
		SplunkTraffic = zap.New(trafficCore, zap.AddCaller()).Sugar()
	} else {
		SplunkTraffic = zap.NewNop().Sugar()
	}

	// --- SPLUNK AUDIT ---
	auditIndex := viper.GetString("splunk.audit_index")
	auditSourceType := viper.GetString("splunk_audit_source")
	if auditSourceType == "" {
		auditSourceType = "netbird:audit"
	}

	if splunkURL != "" && splunkToken != "" && auditIndex != "" {
		auditWriter := NewSplunkHECWriter(splunkURL, splunkToken, auditIndex, splunkSource, auditSourceType, host, 5*time.Second)
		auditCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(auditWriter), zapcore.InfoLevel)
		SplunkAudit = zap.New(auditCore, zap.AddCaller()).Sugar()
	} else {
		SplunkAudit = zap.NewNop().Sugar()
	}

	return nil
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
	if SplunkTraffic != nil {
		_ = SplunkTraffic.Sync()
	}
	if SplunkAudit != nil {
		_ = SplunkAudit.Sync()
	}
}
