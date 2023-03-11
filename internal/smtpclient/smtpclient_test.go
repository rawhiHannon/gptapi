package smtpclient

import (
	"gptapi/pkg/enum"
	"gptapi/pkg/utils"
	"os"
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {
	utils.LoadEnv("/../../")
	// initializing the client
	port, _ := strconv.ParseInt(os.Getenv("SMTP_TEST_PORT"), 10, 32)
	client := New(enum.TLS, os.Getenv("SMTP_TEST_HOST"), int(port), os.Getenv("SMTP_TEST_FROM"), os.Getenv("SMTP_TEST_USER"), os.Getenv("SMTP_TEST_PASS"))
	// check if client is created or not
	if client == nil {
		t.Errorf("Error in creating the client")
	}
}

func TestGetClient(t *testing.T) {
	utils.LoadEnv("/../../")
	port, _ := strconv.ParseInt(os.Getenv("SMTP_TEST_PORT"), 10, 32)
	client := New(enum.TLS, os.Getenv("SMTP_TEST_HOST"), int(port), os.Getenv("SMTP_TEST_FROM"), os.Getenv("SMTP_TEST_USER"), os.Getenv("SMTP_TEST_PASS"))
	c := client.getClient()
	// check if client is created or not
	if c == nil {
		t.Errorf("Error in creating the client")
	}
}

func TestSendMail(t *testing.T) {
	utils.LoadEnv("/../../")
	port, _ := strconv.ParseInt(os.Getenv("SMTP_TEST_PORT"), 10, 32)
	client := New(enum.TLS, os.Getenv("SMTP_TEST_HOST"), int(port), os.Getenv("SMTP_TEST_FROM"), os.Getenv("SMTP_TEST_USER"), os.Getenv("SMTP_TEST_PASS"))
	err := client.sendMail("rawhi.h@pandats.com", "Test Subject", "Test Message")
	// check if there's no error
	if err != nil {
		t.Errorf("Error in sending the mail: %s", err)
	}
}

func TestSend(t *testing.T) {
	utils.LoadEnv("/../../")
	port, _ := strconv.ParseInt(os.Getenv("SMTP_TEST_PORT"), 10, 32)
	client := New(enum.TLS, os.Getenv("SMTP_TEST_HOST"), int(port), os.Getenv("SMTP_TEST_FROM"), os.Getenv("SMTP_TEST_USER"), os.Getenv("SMTP_TEST_PASS"))
	client.Send("rawhi.h@pandats.com", "Test Subject", "Test Message")
}
