package eventdto

import (
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	allowedKinds    = map[string]struct{}{"payment.created": {}}
	allowedStatuses = map[string]struct{}{"pending": {}, "success": {}, "failed": {}}
	allowedCurrency = map[string]struct{}{"RUB": {}, "USD": {}, "EUR": {}}
	allowedDevices  = map[string]struct{}{"card": {}, "apple_pay": {}, "google_pay": {}}
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Errors []FieldError
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%d validation error(s)", len(e.Errors))
}

func (e *ValidationError) Ok() bool {
	return len(e.Errors) == 0
}

func (e *ValidationError) add(field, message string) {
	e.Errors = append(e.Errors, FieldError{Field: field, Message: message})
}

func (r CreateRequest) Validate() error {
	v := &ValidationError{}

	if strings.TrimSpace(r.Kind) == "" {
		v.add("kind", "is required") 
	} else if _, ok := allowedKinds[r.Kind]; !ok {
		v.add("kind", "unsuppoted value")
	}

	if strings.TrimSpace(r.Status) == "" {
		v.add("status", "is required")
	} else if _, ok := allowedStatuses[r.Status]; !ok {
		v.add("status", "unsupported state")
	}

	r.Data.validate(v)

	if !v.Ok() {
		return v
	}

	return nil
}

func (d EventData) validate(v *ValidationError) {
	d.PaymentData.validate(v)
	d.PayDevice.validate(v)
}

func (p PaymentData) validate(v *ValidationError) {
	if strings.TrimSpace(p.AccountID) == "" {
		v.add("data.payment_data.account_id", "is required")
	}
	
	if strings.TrimSpace(p.IP) == "" {
		v.add("data.payment_data.ip_addres", "is required")
	}

	if p.Amount <= 0 {
		v.add("data.payment_data.amount", "must be greater thwn 0")
	}

	if _, ok := allowedCurrency[p.Currency]; !ok {
		v.add("data.payment_data.currency", "unsupported currency")
	}

		if _, ok := allowedStatuses[p.Status]; !ok {
		v.add("data.payment_data.status", "unsupported value")
	}
	if p.Timestamp <= 0 {
		v.add("data.payment_data.timestamp", "must be a unix timestamp")
	} else if time.Unix(p.Timestamp, 0).After(time.Now().Add(24 * time.Hour)) {
		v.add("data.payment_data.timestamp", "is too far in the future")
	}

	if p.UserEmail == "" {
		v.add("data.payment_data.user_email", "is required")
	} else if _, err := mail.ParseAddress(p.UserEmail); err != nil {
		v.add("data.payment_data.user_email", "invalid email")
	}

	if strings.TrimSpace(p.UserPhone) == "" {
		v.add("data.payment_data.user_phone", "is required")
	}

	p.Location.validate(v)
}

func (d PayDevice) validate(v *ValidationError) {
	if strings.TrimSpace(d.Token) == "" {
		v.add("data.payment_device.token", "is required")
	} else if err := uuid.Validate(d.Token); err != nil {
		v.add("data.payment_device.token", "unsupported format")
	}

	if _, ok := allowedDevices[d.Type]; !ok {
		v.add("data.payment_device.type", "unsupported type")
	}
}

func (l Location) validate(v *ValidationError) {
	lat, errLat := strconv.ParseFloat(l.Lat, 64)
	if errLat != nil || lat < -90 || lat > 90 {
		v.add("data.payment_data.location.lat", "must be a number between -90 and 90")
	}

	lon, errLon := strconv.ParseFloat(l.Lon, 64)
	if errLon != nil || lon < -180 || lon > 180 {
		v.add("data.payment_data.location.lon", "must be a number between -180 and 180")
	}
}