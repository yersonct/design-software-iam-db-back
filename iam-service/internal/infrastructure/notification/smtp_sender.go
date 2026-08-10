package notification

import (
	"context"
	"fmt"
	"net/smtp"
)

type SMTPSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPSender(host, port, username, password, from string) *SMTPSender {
	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SMTPSender) SendPasswordResetEmail(ctx context.Context, to string, resetLink string) error {
	subject := "Recuperación de contraseña - IAM SENA"

	body := fmt.Sprintf(
		"Hola,\r\n\r\n"+
			"Recibimos una solicitud para restablecer tu contraseña.\r\n\r\n"+
			"Haz clic en el siguiente enlace (válido por 1 hora):\r\n%s\r\n\r\n"+
			"Si no solicitaste este cambio, ignora este correo.\r\n",
		resetLink,
	)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n%s",
		s.from, to, subject, body,
	))

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}

// SendWelcomeEmail se dispara cuando un administrador crea una cuenta
// nueva (HU: creación de cuentas). Envía la contraseña temporal al
// usuario para que pueda iniciar sesión por primera vez y cambiarla.
func (s *SMTPSender) SendWelcomeEmail(ctx context.Context, to string, firstName string, tempPassword string, loginURL string) error {
	subject := "Tu cuenta en IAM SENA fue creada"

	body := fmt.Sprintf(
		"Hola %s,\r\n\r\n"+
			"Se creó una cuenta para ti en el sistema IAM SENA.\r\n\r\n"+
			"Correo: %s\r\n"+
			"Contraseña temporal: %s\r\n\r\n"+
			"Ingresa aquí y te recomendamos cambiarla apenas inicies sesión:\r\n%s\r\n\r\n"+
			"Si no esperabas este correo, contacta al administrador del sistema.\r\n",
		firstName, to, tempPassword, loginURL,
	)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n%s",
		s.from, to, subject, body,
	))

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}