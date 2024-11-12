package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/Shopify/sarama"
)

func main() {
	// Kafka 서버와 관련된 설정
	brokers := []string{"localhost:9092", "localhost:9093", "localhost:9094"} // Kafka 브로커 주소 리스트
	topic := "my_topic"                                                       // 메시지를 보낼 토픽 이름

	// Sarama 구성 설정
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // 메시지 확인 대기
	config.Producer.Retry.Max = 5                    // 재시도 횟수
	config.Producer.Return.Successes = true          // 성공한 메시지 반환

	// SSL 설정
	config.Net.TLS.Enable = true

	// CA 인증서 로드
	caCert, err := ioutil.ReadFile("../../tmp/datahub-ca.crt") // CA 인증서 경로
	if err != nil {
		log.Fatalf("CA 인증서를 읽는 중 오류 발생: %s", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatal("CA 인증서를 추가할 수 없습니다.")
	}

	// 클라이언트 인증서 및 키 로드
	cert, err := tls.LoadX509KeyPair("../../secrets/localhost.crt.pem", "../../secrets/localhost.key.pem") // 클라이언트 인증서 및 키 경로
	if err != nil {
		log.Fatalf("클라이언트 인증서를 로드하는 중 오류 발생: %s", err)
	}

	config.Net.TLS.Config = &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true, // 인증서 확인을 건너뛸지 여부 설정
	}

	// Sarama 프로듀서 생성
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Kafka 프로듀서를 생성하는 중 오류 발생: %s", err)
	}
	defer producer.Close()

	// 전송할 메시지 생성
	message := "Hello, Kafka with SSL!" // 보낼 메시지 내용
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	// 메시지 전송
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Fatalf("메시지를 전송하는 중 오류 발생: %s", err)
	}

	fmt.Printf("메시지 전송 성공! Partition: %d, Offset: %d\n", partition, offset)
}
