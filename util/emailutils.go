package util

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"net"
	"net/smtp"

	"github.com/merinovvvv/momentic-backend/initializers"
	"github.com/merinovvvv/momentic-backend/models"
	// "gorm.io/gorm"
)

func SendVerificationEmail(receiver, code string) error {
	sender := strings.TrimSpace(os.Getenv("EMAIL"))
	senderPassword := os.Getenv("EMAIL_PASSWORD")
	if sender == "" || senderPassword == "" {
		return errors.New("missing EMAIL or EMAIL_PASSWORD environment variable")
	}
	receiver = strings.TrimSpace(receiver)
	if receiver == "" {
		return errors.New("receiver email is empty")
	}

	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587"
	}
	addr := net.JoinHostPort(smtpHost, smtpPort)

	subject := fmt.Sprintf("Your Momentic code is %s", code)
	body := "If you haven't registred a Momentic account, ignore this email"

	msg := "From: " + sender + "\n" +
		"To: " + receiver + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial %s: %w", addr, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer func() {
		_ = client.Quit()
	}()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: smtpHost}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	auth := smtp.PlainAuth("", sender, senderPassword, smtpHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	// send mail transaction
	if err := client.Mail(sender); err != nil {
		return fmt.Errorf("mail from=%s: %w", sender, err)
	}
	if err := client.Rcpt(receiver); err != nil {
		return fmt.Errorf("rcpt to=%s: %w", receiver, err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		_ = w.Close()
		return fmt.Errorf("write msg: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("finish message: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
    
	// TODO: hash code
	emailVerification := models.EmailVerification {
		Email: receiver,
		Code: code,
		ExpiresAt: time.Now().Add(time.Minute * 15),
	}
	if err := initializers.DB.Create(&emailVerification).Error; err != nil {
		return fmt.Errorf("failed to create email verification: %w", err)
	}
	// if err := initializers.DB.First(&emailVerification, "email = ?", receiver).Error; err != nil {
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		if err := initializers.DB.Create(&emailVerification).Error; err != nil {
	// 			return fmt.Errorf("failed to create email verification: %w", err)
	// 		}
	// 	} else {
	// 		return fmt.Errorf("failed to write email verification to db: %w", err)
	// 	}
	// }
	log.Printf("Email sent from %s to %s", sender, receiver)
	return nil
}

// Generates a random number of given length
func GenerateVerificationCode(length int) (error, string) {
	if length < 1 || length > 9 {
		return errors.New("length must be between 1 and 9"), ""
	}
    var max int
    for i := range length {
        max += 9*int((math.Pow(10, float64(i))))
    }
	
	return nil, fmt.Sprintf("%0*d",length, rand.Intn(max))
}
