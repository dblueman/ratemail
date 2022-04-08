package ratemail

import (
   "regexp"
   "sync"
   "time"

   "github.com/Shopify/gomail"
)

const (
   expiry = time.Hour * 24
)

var (
   cache = map[string]time.Time{}
   lock  = sync.Mutex{}
   re    = regexp.MustCompile(`\d{2}:\d{2}:\d{2}`)
)

type Mailer struct {
   *gomail.Dialer
   from string
}

func init() {
   go func() {
      time.Sleep(expiry)

      lock.Lock()
      defer lock.Unlock()

      for key, t0 := range(cache) {
         if time.Since(t0) < expiry {
            continue
         }

         delete(cache, key)
      }
   }()
}

func NewMailer(host string, port int, user, pass, from string) *Mailer {
   mailer := Mailer{
      gomail.NewDialer(host, port, user, pass),
      from,
   }

   mailer.Timeout = 60 * time.Second
   return &mailer
}

func (mailer *Mailer) Check() error {
   s, err := mailer.Dial()
   if err == nil {
      s.Close()
   }
   return err
}

func (mailer *Mailer) Send(to, subject, bodyType, body string) error {
   msg := gomail.NewMessage()
   msg.SetHeader("From", mailer.from)
   msg.SetHeader("To", to)
   msg.SetHeader("Subject", subject)
   msg.SetBody(bodyType, body)

   return mailer.DialAndSend(msg)
}

func (mailer *Mailer) SendRate(to, subject, bodyType, body string) error {
   key := re.ReplaceAllString(to+subject+body, "")

   lock.Lock()
   defer lock.Unlock()

   _, ok := cache[key]
   if ok {
      return nil
   }

   ret := mailer.Send(to, subject, bodyType, body)

   if ret == nil {
      cache[key] = time.Now()
   }

   return ret
}
