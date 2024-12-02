package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"os"
	"testing"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"
	mock "ya-prac-project1/internal/services/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetRuntimeMetrics(t *testing.T) {
	ms := getRuntimeMetrics()

	assert.Equal(t, 30, len(ms))
}

func TestGetPoolCountMetric(t *testing.T) {
	m, _ := getPoolCountMetric()

	expect := metrics.NewMetric("PollCount", metrics.MetricTypeCounter, "1")
	assert.Equal(t, expect, m)
}

func TestGetRandomValueMetirc(t *testing.T) {
	m, _ := getRandomValueMetirc()

	assert.NotEqual(t, "", m.GetValue())
}

func TestUpdateRuntimeMetrics(t *testing.T) {
	logger.Set()
	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	store.EXPECT().SetMetrics(gomock.Any()).AnyTimes()

	ctx, stop := context.WithTimeout(context.Background(), 1*time.Second)
	s := NewRuntimeService(store)
	s.Run(ctx, 10)
	<-ctx.Done()
	stop()
}

func TestUpdateRuntimeMetrics2(t *testing.T) {
	logger.Set()
	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	store.EXPECT().GetMetrics().Return([]metrics.Metrics{}).AnyTimes()
	store.EXPECT().SetMetrics(gomock.Any()).AnyTimes()
	ctx, stop := context.WithTimeout(context.Background(), 2*time.Second)

	s := NewRuntimeService(store)
	s.updateRuntimeMetrics(ctx, 1)
	<-ctx.Done()
	stop()
}

func TestRunSendRequest(t *testing.T) {
	logger.Set()
	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	store.EXPECT().GetMetrics().Return([]metrics.Metrics{}).AnyTimes()

	s := NewRuntimeService(store)
	rCh := make(chan *http.Request, 1)
	s.RunSendRequest(rCh, "", "", "")
	<-rCh
}

func TestRunSendRequest2(t *testing.T) {
	logger.Set()
	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	store.EXPECT().GetMetrics().Return([]metrics.Metrics{}).AnyTimes()

	s := NewRuntimeService(store)
	rCh := make(chan *http.Request, 1)
	s.RunSendRequest(rCh, "", "test", "")
	<-rCh
}

func TestRunSendgRPCRequest_1(t *testing.T) {
	logger.Set()
	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	store.EXPECT().GetMetrics().Return([]metrics.Metrics{}).AnyTimes()

	s := NewRuntimeService(store)
	s.RunSendgRPCRequest("")
}

// func TestEncryptMessage(t *testing.T) {
// 	encryptMessage(*bytes.NewBuffer([]byte{}), "")
// }
// func TestEncryptMessage2(t *testing.T) {
// 	encryptMessage(*bytes.NewBuffer([]byte{}), "testdata/public.pem")
// }

// func TestEncryptMessage3(t *testing.T) {
// 	encryptMessage(*bytes.NewBuffer([]byte{}), "testdata/wrong_public.pem")
// }

func TestEncryptMessage_NoCryptoKey(t *testing.T) {
	// Входной буфер
	input := bytes.NewBufferString("test message")

	// Вызываем функцию с пустым ключом
	output := encryptMessage(*input, "")

	// Проверяем, что выходной буфер идентичен входному
	assert.Equal(t, input.Bytes(), output.Bytes())
}

func TestEncryptMessage_ReadKeyError(t *testing.T) {
	// Входной буфер
	input := bytes.NewBufferString("test message")

	// Вызываем функцию с несуществующим ключом
	output := encryptMessage(*input, "/nonexistent/path/to/key")

	// Проверяем, что выходной буфер идентичен входному
	assert.Equal(t, input.Bytes(), output.Bytes())
}

func TestEncryptMessage_InvalidPEMBlock(t *testing.T) {
	// Создаем временный файл с некорректным PEM-блоком
	tempFile, err := os.CreateTemp("", "invalid_pem")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("INVALID PEM BLOCK")
	assert.NoError(t, err)

	// Входной буфер
	input := bytes.NewBufferString("test message")

	// Вызываем функцию
	output := encryptMessage(*input, tempFile.Name())

	// Проверяем, что выходной буфер идентичен входному
	assert.Equal(t, input.Bytes(), output.Bytes())
}

func TestEncryptMessage_NonRSAPublicKey(t *testing.T) {
	// Создаем временный файл с некорректным типом ключа
	tempFile, err := os.CreateTemp("", "non_rsa_key")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	block := &pem.Block{
		Type:  "EC PUBLIC KEY",
		Bytes: []byte("INVALID KEY DATA"),
	}
	err = pem.Encode(tempFile, block)
	assert.NoError(t, err)

	// Входной буфер
	input := bytes.NewBufferString("test message")

	// Вызываем функцию
	output := encryptMessage(*input, tempFile.Name())

	// Проверяем, что выходной буфер идентичен входному
	assert.Equal(t, input.Bytes(), output.Bytes())
}

func TestEncryptMessage_ParseKeyError(t *testing.T) {
	// Создаем временный файл с некорректным ключом
	tempFile, err := os.CreateTemp("", "invalid_key")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: []byte("INVALID KEY DATA"),
	}
	err = pem.Encode(tempFile, block)
	assert.NoError(t, err)

	// Входной буфер
	input := bytes.NewBufferString("test message")

	// Вызываем функцию
	output := encryptMessage(*input, tempFile.Name())

	// Проверяем, что выходной буфер идентичен входному
	assert.Equal(t, input.Bytes(), output.Bytes())
}

func TestEncryptMessage_EncryptTooLong(t *testing.T) {
	// Создаем RSA ключи
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	publicKey := &privateKey.PublicKey

	// Создаем PEM-блок для публичного ключа
	pubKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	// Создаем временный файл для публичного ключа
	tempFile, err := os.CreateTemp("", "public_key")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	err = pem.Encode(tempFile, block)
	assert.NoError(t, err)

	// Генерируем слишком длинное сообщение
	tooLongMessage := make([]byte, 300) // Длина превышает лимит для 2048-битного ключа
	for i := range tooLongMessage {
		tooLongMessage[i] = byte('a')
	}
	input := bytes.NewBuffer(tooLongMessage)

	// Вызываем функцию
	output := encryptMessage(*input, tempFile.Name())

	// Проверяем, что выходной буфер идентичен входному (шифрование не произошло)
	assert.Equal(t, input.Bytes(), output.Bytes())
}

func TestEncryptMessage_EncryptionSuccess(t *testing.T) {
	// Создаем RSA ключи
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	publicKey := &privateKey.PublicKey

	// Создаем PEM-блок для публичного ключа
	pubKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	// Создаем временный файл для публичного ключа
	tempFile, err := os.CreateTemp("", "public_key")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	err = pem.Encode(tempFile, block)
	assert.NoError(t, err)

	// Входной буфер
	input := bytes.NewBufferString("test message")

	// Вызываем функцию
	output := encryptMessage(*input, tempFile.Name())

	// Проверяем, что выходной буфер отличается от входного (шифрование произошло)
	assert.NotEqual(t, input.Bytes(), output.Bytes())

	// Дешифруем сообщение для проверки
	decryptedMessage, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, output.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, input.Bytes(), decryptedMessage)
}
