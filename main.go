package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func (envelop Envelope) getClaim() Email {
	return envelop.Body.SendEmailServiceReq.Email
}

// Defina as estruturas de dados necessárias para o SOAP Envelope, Body e sua carga útil.
type Envelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    Body     `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

type Body struct {
	SendEmailServiceReq SendEmailServiceReq `xml:"http://www.smartnet.com.br/services/esbEmailService SendEmailService_Req"`
}

type SendEmailServiceReq struct {
	Email Email `xml:"email"`
}

type Email struct {
	ToAddress        string `xml:"toAddress"`
	FromAddress      string `xml:"fromAddress"`
	SubjectOfEmail   string `xml:"subjectOfEmail"`
	EmailMessageText string `xml:"emailMessageText"`
}

// Manipula solicitações SOAP
func handleSOAPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are supported", http.StatusMethodNotAllowed)
		return
	}

	// Leia o corpo da solicitação
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// Deserializar o corpo SOAP no struct Envelope
	var envelope Envelope
	err = xml.Unmarshal(body, &envelope)
	if err != nil {
		http.Error(w, "Error unmarshalling XML", http.StatusBadRequest)
		return
	}

	// Agora você pode usar a variável 'envelope' conforme necessário.
	// Exemplo: Imprimir o conteúdo do envelope
	fmt.Println("Envia e-mail em background fila...")
	go sendEmail(envelope.getClaim())

	// Responder ao cliente
	_, err = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?><soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
	<soapenv:Body><NS1:SendEmailService_Resp xmlns:NS1="http://www.smartnet.com.br/services/esbEmailService">
	<return><codRet>-99</codRet><msgRet>Erro no serviço: SOAP Envelope has invalid namespace--Envelope</msgRet>
	</return></NS1:SendEmailService_Resp></soapenv:Body></soapenv:Envelope>`))
	if err != nil {
		fmt.Println(err)
	}
}

func sendEmail(email Email) {

	// Configurar o corpo do email
	message := fmt.Sprintf("Subject: %s\r\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s",
		email.SubjectOfEmail, email.EmailMessageText)

	// Configurar autenticação
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpServer)

	// Conectar ao servidor SMTP
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", smtpServer, smtpPort),
		auth,
		email.FromAddress,
		[]string{email.ToAddress},
		[]byte(message),
	)

	if err != nil {
		fmt.Println("Erro ao enviar o email:", err)
		return
	}

	fmt.Println("Email enviado com sucesso!")
}

func main() {
	// Roteamento de solicitações SOAP para a função de manipulação
	http.HandleFunc("/", handleSOAPRequest)

	// Inicie o servidor na porta srvPort
	fmt.Printf("Starting server on :%d...\n", int32(srvPort))
	err := http.ListenAndServe(fmt.Sprintf(":%d", srvPort), nil)
	if err != nil {
		fmt.Println(err)
	}
}

var smtpPassword string

var smtpServer string

var smtpPort string
var smtpUser string
var srvPort int
var srvPath string = "/"

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env:", err)
	}

	smtpPassword = os.Getenv("SMTP_PASSWORD")
	smtpServer = os.Getenv("SMTP_SERVER")
	smtpPort = os.Getenv("SMTP_PORT")
	smtpUser = os.Getenv("SMTP_USER")
	srvPort, _ = strconv.Atoi(os.Getenv("HTTP_SERVER_PORT"))
	srvPath = os.Getenv("HTTP_SERVER_PATH")

}
