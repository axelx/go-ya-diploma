package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/axelx/go-ya-diploma/internal/handlers"
	"github.com/axelx/go-ya-diploma/internal/logger"
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/axelx/go-ya-diploma/internal/orderservice"
	"github.com/axelx/go-ya-diploma/internal/pg"
	"github.com/axelx/go-ya-diploma/internal/userservice"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

var (
	DBTestURL = "postgres://user:password@localhost:5464/go-ya-gophermart-test"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body []byte) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader(body))

	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
func testRequestAuth(t *testing.T, ts *httptest.Server, method, path string, body []byte, userID string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader(body))

	req.AddCookie(&http.Cookie{Name: "auth", Value: userID, Expires: time.Now().Add(time.Hour * 1)})

	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestAddOrder(t *testing.T) {
	fmt.Println("--TestAddOrder", `{"login": "qwerty833316","password": "password"}`)

	chNewOrder := make(chan string, 100)
	lg := logger.Initialize("info")
	db, _ := pg.InitDB(DBTestURL, lg)
	//pg.DropTablesForTest(db, lg)

	ord := orderservice.Order{DB: db, LG: lg}
	usr := userservice.User{DB: db, LG: lg}

	orderNum := orderservice.GenerateLunaNumber(11)

	hd := handlers.New(ord, usr, lg, db, chNewOrder)
	ts := httptest.NewServer(hd.Router())

	userID := findOrCreateUser(db, lg, "test")

	defer ts.Close()

	var testTable = []struct {
		url    string
		want   string
		status int
		body   string
	}{
		{
			url:    "/api/user/orders",
			want:   ``,
			status: http.StatusAccepted,
			body:   orderNum,
		},
		{
			url:    "/api/user/orders",
			want:   ``,
			status: http.StatusOK,
			body:   orderNum,
		},
	}

	for _, v := range testTable {
		resp, data := testRequestAuth(t, ts, "POST", v.url, []byte(v.body), strconv.Itoa(userID))

		var result models.User
		json.Unmarshal([]byte(data), &result)

		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode)
		assert.Equal(t, v.want, data)
	}

}

func TestCreateUser(t *testing.T) {
	chNewOrder := make(chan string, 100)
	lg := logger.Initialize("info")
	db, _ := pg.InitDB(DBTestURL, lg)

	ord := orderservice.Order{DB: db, LG: lg}
	usr := userservice.User{DB: db, LG: lg}

	hd := handlers.New(ord, usr, lg, db, chNewOrder)

	ts := httptest.NewServer(hd.Router())
	login := randomString(5)

	defer ts.Close()

	var testTable = []struct {
		url    string
		want   string
		status int
		body   string
	}{
		{
			url:    "/api/user/register",
			want:   `{"login":"` + login + `", "password":"password"}`,
			status: http.StatusOK,
			body:   `{"login":"` + login + `", "password":"password"}`,
		},
		{
			url: "/api/user/register",
			want: `StatusConflict
`,
			status: http.StatusConflict,
			body:   `{"login":"` + login + `", "password":"password"}`,
		},
	}

	for _, v := range testTable {
		resp, data := testRequest(t, ts, "POST", v.url, []byte(v.body))

		var result models.User
		json.Unmarshal([]byte(data), &result)

		fmt.Println(result)

		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode)
		assert.Equal(t, v.want, data)
	}
}

func TestCheckAccural(t *testing.T) {
	urlAccrualServer := ""
	order := "91"
	chProcOrder := make(chan string, 100)
	chAddOrder := make(chan string, 100)
	countPerMin := 0

	lg := logger.Initialize("info")
	db, err := pg.InitDB(DBTestURL, lg)

	o := orderservice.Order{DB: db, LG: lg}
	usr := userservice.User{DB: db, LG: lg}
	usr.Create("test1", "test1")
	usrID, _ := usr.SearchOne("test1")
	o.Create(usrID, order, 0, chAddOrder)

	if err != nil {
		lg.Error("Error not connect to pg", zap.String("about ERR", err.Error()))
	}

	accrualServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("--011", r.URL.String(), "order", order)
		assert.Equal(t, urlAccrualServer+"/api/orders/"+order, r.URL.String())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"PROCESSED","accrual":10.5, "order":` + order + `}`))
	}))
	defer accrualServer.Close()
	checkAccural(accrualServer.URL+"/", "91", chProcOrder, &countPerMin, db, lg)

	findOrder, _ := o.SearchMany(usrID)
	assert.Equal(t, 10.5, findOrder[0].Accrual) // Проверяем, что countPerMin увеличился
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}

func findOrCreateUser(db *sqlx.DB, lg *zap.Logger, login string) int {

	userID, _ := pg.FindUserByLogin(db, lg, login)

	if userID == 0 {
		pg.CreateNewUser(db, lg, login, "111")
		userID, _ = pg.FindUserByLogin(db, lg, login)

	}
	return userID
}
