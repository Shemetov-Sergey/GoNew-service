package postgres

import (
	"GoNew-service/pkg/storage"
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func GetTestDb() (*Storage, error) {
	e := godotenv.Load("../../../.env") //Загрузить файл .env
	if e != nil {
		log.Print(e)
		return nil, e
	}

	username := os.Getenv("db_user")
	password := os.Getenv("db_pass")
	dbName := os.Getenv("test_db_name")
	dbHost := os.Getenv("db_host")
	dbPort := os.Getenv("db_port")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, dbHost, dbPort, dbName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	postsChan := make(chan storage.Post)
	errChan := make(chan error)

	db, err := New(ctx, connString, postsChan, errChan)

	if err != nil {
		log.Fatal(err)
		return nil, e
	}

	return db, nil
}

func TestStorage_AddPost(t *testing.T) {
	db, err := GetTestDb()

	if err != nil {
		t.Fatal(err)
	}

	p := storage.Post{
		Title: "Аналитики оценили вывод средств с криптобиржи Binance за сутки в $1,9 млрд",
		Content: "Пользователи за сутки вывели с криптобиржи Binance $1,9 млрд, " +
			"сообщила компания по анализу блокчейн-данных Nansen. В понедельник, 12 декабря, " +
			"издание CoinDesk оценило отток средств с Binance в $902 млн. Во вторник, 13 декабря, " +
			"Binance на время приостанавливала вывод средств в стэйблкоинах USDC, " +
			"объяснив это необходимостью взаимодействовать с банками в Нью-Йорке и задержками при обмене токенов",
		PubTime:       1670979034,
		Link:          "https://www.forbes.ru/investicii/482462-analitiki-ocenili-vyvod-sredstv-s-kriptobirzi-binance-za-sutki-v-1-9-mlrd",
		SourceXmlLink: "http://www.forbes.ru/newrss.xml",
	}

	e := db.AddPost(p)

	if e != nil {
		t.Fatal(err)
	}
}

func TestStorage_GetLastPubDateForSources(t *testing.T) {
	db, err := GetTestDb()

	if err != nil {
		t.Fatal(err)
	}

	sourceSlice := []string{"https://habr.com/ru/rss/hub/go/all/?fl=ru?limit=10"}

	// Значение не проверяем так как может быть 0 и более
	_, err = db.GetLastPubDateForSources(sourceSlice)

	if err != nil {
		t.Fatal(err)
	}
}

func TestStorage_Posts(t *testing.T) {
	db, err := GetTestDb()

	if err != nil {
		t.Fatal(err)
	}

	got, err := db.Posts(10)

	if len(got) != 10 {
		t.Errorf("Posts() len = %v, want %v", len(got), 10)
	}

	if err != nil {
		t.Errorf("Posts() error = %v, wantErr %v", err, nil)
		return
	}
}
