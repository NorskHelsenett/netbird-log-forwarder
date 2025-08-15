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

var Splunk *zap.SugaredLogger
var Log *zap.SugaredLogger

type SplunkConfig struct {
	Url   string
	Token string
}

func InitLogger(logDir string) error {

	// --- Splunk ---
	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "level",
		MessageKey:   "message",
		CallerKey:    "caller",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	})

	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "level",
		MessageKey:   "message",
		CallerKey:    "caller",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.CapitalColorLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	})

	splunkURL := viper.GetString("splunk_url")
	splunkToken := viper.GetString("splunk_token")
	splunkIndex := viper.GetString("splunk_index")
	splunkSource := "aws.s3"
	splunkSourceType := "netbird:traffic"

	splunkWriter := NewSplunkHECWriter(
		splunkURL,
		splunkToken,
		splunkIndex,
		splunkSource,
		splunkSourceType,
		5*time.Second,
	)

	splunkCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(splunkWriter), zapcore.InfoLevel)
	Splunk = zap.New(splunkCore, zap.AddCaller()).Sugar()

	// --- CONSOLE & FILE ---

	// --- File core ---
	// file, err := os.OpenFile(logDir+"/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	return err
	// }

	lumberJack := &lumberjack.Logger{
		Filename:   logDir + "/app.log",
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}

	fileCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(lumberJack), zapcore.InfoLevel)

	// --- Console core ---
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)

	// --- Samlet App-logger til fil og konsoll ---
	appCore := zapcore.NewTee(fileCore, consoleCore)
	Log = zap.New(appCore, zap.AddCaller()).Sugar()

	return nil
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
	if Splunk != nil {
		_ = Splunk.Sync()
	}
}

// func getWriter(path string) *os.File {
// 	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		panic("unable to open log file: " + err.Error())
// 	}
// 	return f
// }

type SplunkHECWriter struct {
	URL        string
	Token      string
	Index      string
	Source     string
	SourceType string
	Client     *resty.Client
}

func NewSplunkHECWriter(url, token, index, source, sourcetype string, timeout time.Duration) *SplunkHECWriter {
	return &SplunkHECWriter{
		URL:        url,
		Token:      token,
		Index:      index,
		Source:     source,
		SourceType: sourcetype,
		Client: resty.New().
			SetTimeout(timeout),
	}
}

func (w *SplunkHECWriter) Write(p []byte) (n int, err error) {
	var logTimestamp int64
	var logEntry map[string]interface{}

	// Unmarshal original Zap event
	if err := json.Unmarshal(p, &logEntry); err != nil {
		logTimestamp = time.Now().Unix()
	} else {
		// Extract timestamp
		if tsStr, ok := logEntry["time"].(string); ok {
			t, err := time.Parse(time.RFC3339, tsStr)
			if err == nil {
				logTimestamp = t.Unix()
			} else {
				logTimestamp = time.Now().Unix()
			}
		} else {
			logTimestamp = time.Now().Unix()
		}

		// Remove unwanted fields from event payload
		delete(logEntry, "time")
		delete(logEntry, "timestamp")
		delete(logEntry, "caller")
	}

	// Build Splunk payload
	payload := map[string]interface{}{
		"event":      logEntry,
		"time":       logTimestamp,
		"host":       "netbird-test-logger",
		"source":     w.Source,
		"sourcetype": w.SourceType,
		"index":      w.Index,
	}

	prettyPayload, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Println("Sending to Splunk:", string(prettyPayload))

	// Send to Splunk HEC using Resty
	resp, err := w.Client.R().
		SetHeader("Authorization", "Splunk "+w.Token).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(w.URL + "/services/collector/event")

	// Log response from Splunk
	if resp.StatusCode() != 200 {
		fmt.Println("Failed to send to Splunk:", resp.Status())
	}

	if err != nil {
		// Fail silently — do not block app
		return len(p), nil
	}

	if resp.StatusCode() != 200 {
		// Optionally log here
		return len(p), nil
	}

	return len(p), nil
}
