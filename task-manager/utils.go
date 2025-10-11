package taskmanager

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

func ValidateCrontab(crontab string) bool {
	p := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	withLocation := fmt.Sprintf("CRON_TZ=%s %s", time.UTC.String(), crontab)
	_, err := p.Parse(withLocation)
	return err == nil
}
