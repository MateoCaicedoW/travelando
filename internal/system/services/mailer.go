package services

import (
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"time"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
)

func SendReservationConfirmation(inquiry models.Inquiry) error {
	host := os.Getenv("SMTP_SERVER")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")
	if host == "" || port == "" || username == "" || password == "" || from == "" {
		return errors.New("SMTP is not configured")
	}
	fromAddress, err := mail.ParseAddress(from)
	if err != nil || fromAddress.Address != from {
		return errors.New("SMTP_FROM must be a valid email address")
	}
	toAddress, err := mail.ParseAddress(inquiry.GuestEmail)
	if err != nil || toAddress.Address != inquiry.GuestEmail {
		return errors.New("guest email is invalid")
	}

	addr := net.JoinHostPort(host, port)
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}
	tlsConn := tls.Client(conn, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err := tlsConn.Handshake(); err != nil {
		_ = conn.Close()
		return fmt.Errorf("start SMTP TLS: %w", err)
	}
	client, err := smtp.NewClient(tlsConn, host)
	if err != nil {
		_ = tlsConn.Close()
		return fmt.Errorf("open SMTP client: %w", err)
	}
	defer client.Close()

	if err := client.Auth(smtp.PlainAuth("", username, password, host)); err != nil {
		return fmt.Errorf("authenticate with SMTP server: %w", err)
	}
	if err := client.Mail(fromAddress.Address); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(toAddress.Address); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("open SMTP message: %w", err)
	}

	stay := helpers.StayTypeLabel(inquiry.StayType)
	dateText := helpers.DateLabel(inquiry.StayType, inquiry.DateFrom, inquiry.DateTo)
	subject := mime.QEncoding.Encode("UTF-8", "Reserva confirmada | Travelando")
	body := fmt.Sprintf("Hola %s,\n\nTu reserva quedó confirmada.\n\nCabaña: %s\nTipo: %s\nFechas: %s\nPersonas: %d\nTotal acordado: $%s COP\n\nEl pago se realiza directamente con el contacto de la cabaña.\n\nTravelando\n", inquiry.GuestName, inquiry.CabinName, stay, dateText, inquiry.Guests, helpers.FormatCOP(inquiry.FinalTotal))
	message := fmt.Sprintf("%s\r\n%s\r\n%s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		"From: "+fromAddress.String(),
		"To: "+toAddress.String(),
		"Subject: "+subject,
		body,
	)
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("send SMTP message: %w", err)
	}
	return client.Quit()
}
