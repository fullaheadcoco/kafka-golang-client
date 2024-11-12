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
	topic := "my_topic"                                                       // 소비할 토픽 이름

	// Sarama 구성 설정
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_1_0_0 // Kafka 버전에 맞춰 설정

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

	// Sarama 소비자 생성
	client, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Fatalf("Kafka 소비자를 생성하는 중 오류 발생: %s", err)
	}
	defer client.Close()

	// Kafka 토픽 소비
	partitionConsumer, err := client.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Kafka 파티션을 소비하는 중 오류 발생: %s", err)
	}
	defer partitionConsumer.Close()

	fmt.Println("Kafka 메시지를 수신 대기 중...")

	// 메시지 수신 루프
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			fmt.Printf("메시지 수신: %s\n", string(msg.Value))
		case err := <-partitionConsumer.Errors():
			fmt.Printf("소비자 오류: %s\n", err)
		}
	}
}
