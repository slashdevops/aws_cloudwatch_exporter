package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/common/log"
)

func Test_GetTimeStamps(t *testing.T) {
	type args struct {
		now time.Time
		p   string
	}
	tests := []struct {
		name          string
		args          args
		wantStartTime time.Time
		wantEndTime   time.Time
		wantPeriod    time.Duration
	}{
		{
			name: "Test5mPeriodExact",
			args: args{
				now: parseDate("2020-05-10T11:05:00Z", time.RFC3339),
				p:   "5m",
			},
			wantStartTime: parseDate("2020-05-10T11:00:00Z", time.RFC3339),
			wantEndTime:   parseDate("2020-05-10T11:10:00Z", time.RFC3339),
			wantPeriod:    parseDuration("5m"),
		},
		{
			name: "Test5mPeriodBefore",
			args: args{
				now: parseDate("2020-05-10T11:04:59Z", time.RFC3339),
				p:   "5m",
			},
			wantStartTime: parseDate("2020-05-10T10:55:00Z", time.RFC3339),
			wantEndTime:   parseDate("2020-05-10T11:05:00Z", time.RFC3339),
			wantPeriod:    parseDuration("5m"),
		},
		{
			name: "Test5mPeriodAfter",
			args: args{
				now: parseDate("2020-05-10T11:05:59Z", time.RFC3339),
				p:   "5m",
			},
			wantStartTime: parseDate("2020-05-10T11:00:00Z", time.RFC3339),
			wantEndTime:   parseDate("2020-05-10T11:10:00Z", time.RFC3339),
			wantPeriod:    parseDuration("5m"),
		},
		{
			name: "Test5mPeriodEndDay",
			args: args{
				now: parseDate("2020-05-11T00:04:59Z", time.RFC3339),
				p:   "5m",
			},
			wantStartTime: parseDate("2020-05-10T23:55:00Z", time.RFC3339),
			wantEndTime:   parseDate("2020-05-11T00:05:00Z", time.RFC3339),
			wantPeriod:    parseDuration("5m"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStartTime, gotEndTime, gotPeriod := GetTimeStamps(tt.args.now, tt.args.p)
			if gotStartTime != tt.wantStartTime {
				t.Errorf("GetTimeStamps() gotStartTime = %v, want %v", gotStartTime, tt.wantStartTime)
			}
			if gotEndTime != tt.wantEndTime {
				t.Errorf("GetTimeStamps() gotEndTime = %v, want %v", gotEndTime, tt.wantEndTime)
			}
			if gotPeriod != tt.wantPeriod {
				t.Errorf("GetTimeStamps() gotPeriod = %v, want %v", gotPeriod, tt.wantPeriod)
			}
		})
	}
}

func parseDuration(d string) time.Duration {
	td, err := time.ParseDuration(d)
	if err != nil {
		log.Errorf("Error parsing period: %v, %v", d, err)
	}
	return td
}

func parseDate(d string, l string) time.Time {
	td, err := time.Parse(l, d)
	if err != nil {
		log.Errorf("Error parsing date: %v, %v", d, err)
	}
	return td
}
