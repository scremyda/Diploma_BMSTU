package scheduler

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"diploma/certer/certer"
	"diploma/certer/consumer"
	"diploma/certer/producer"
	"diploma/certer/setter"
	"diploma/models"
)

type Config struct {
	Interval time.Duration `yaml:"interval"`
}

type Scheduler struct {
	consumer consumer.Interface
	producer producer.Interface
	certer   certer.Interface
	setter   setter.Interface
	conf     Config
}

func New(
	conf Config,
	consumer consumer.Interface,
	producer producer.Interface,
	certer certer.Interface,
	setter setter.Interface,
) *Scheduler {
	return &Scheduler{
		consumer: consumer,
		producer: producer,
		certer:   certer,
		setter:   setter,
		conf:     conf,
	}
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	ticker := time.NewTicker(s.conf.Interval)
	defer ticker.Stop()

	if err := s.runOnce(ctx); err != nil {
		log.Println("initial run failed:", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping scheduler")
			return ctx.Err()
		case <-ticker.C:
			if err := s.runOnce(ctx); err != nil {
				log.Println("scheduled run failed:", err)
			}
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context) error {
	events, err := s.consumer.GetCerterEvents(ctx)
	if err != nil {
		return fmt.Errorf("getting events: %w", err)
	}

	var wg sync.WaitGroup
	for _, ev := range events {
		wg.Add(1)
		go s.handleCerterEvent(ctx, ev, &wg)
	}
	wg.Wait()
	return nil
}

func (s *Scheduler) handleCerterEvent(ctx context.Context, ev models.CerterEvent, wg *sync.WaitGroup) {
	defer wg.Done()

	u, err := url.Parse(ev.Target)
	if err != nil {
		log.Printf("invalid URL %q: %v", ev.Target, err)
		return
	}

	certPEM, keyPEM, err := s.certer.GenerateCertSignedByCA(u.Scheme)
	if err != nil {
		log.Printf("failed to generate certs for %q: %v", u.Scheme, err)
		return
	}

	if err := s.setter.Set(ctx, u.Scheme, certPEM, keyPEM); err != nil {
		log.Printf("failed to set certs for %q: %v", u.Scheme, err)
		return
	}

	notBefore, notAfter, issuer := parseCertInfo([]byte(certPEM))

	msg := formatAlertMessage(u.Scheme, issuer, notBefore, notAfter)

	alert := models.AlerterEvent{
		Target:  ev.Target,
		Message: msg,
	}
	if err := s.producer.Produce(ctx, alert); err != nil {
		log.Printf("failed to produce alert for %q: %v", ev.Target, err)
	}
}

func parseCertInfo(certPEM []byte) (notBefore, notAfter, issuer string) {
	notBefore, notAfter, issuer = "unknown", "unknown", "unknown"
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return
	}
	parsed, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return
	}
	notBefore = parsed.NotBefore.Format("02 Jan 2006 15:04")
	notAfter = parsed.NotAfter.Format("02 Jan 2006 15:04")
	issuer = parsed.Issuer.CommonName
	return
}

func formatAlertMessage(service, issuer, notBefore, notAfter string) string {
	return fmt.Sprintf(
		"🔔 Сертификат обновлён! 🔔\n\n"+
			"• Сервис: `%s`\n"+
			"• Издатель: %s\n"+
			"• Действует с: %s\n"+
			"• Действует до: %s\n",
		service, issuer, notBefore, notAfter,
	)
}
