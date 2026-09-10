// hello is the dogfood fixture: it proves the platform injected its identity,
// its secret (by digest only), and a working DATABASE_URL.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

const version = "v1"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })
	http.HandleFunc("/whoami", whoami)
	http.HandleFunc("/admin/", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "admin area\n") })
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "hello %s from %s (release %s)\n", version, os.Getenv("DEPLOYMIND_APP"), os.Getenv("DEPLOYMIND_RELEASE"))
	})
	http.HandleFunc("/env", func(w http.ResponseWriter, _ *http.Request) {
		names := []string{}
		for _, n := range []string{"DEPLOYMIND_APP", "DEPLOYMIND_ORG", "DEPLOYMIND_RELEASE", "DEPLOYMIND_IDENTITY_KEY", "DATABASE_URL", "PORT", "GREETING_SECRET"} {
			if os.Getenv(n) != "" {
				names = append(names, n)
			}
		}
		sort.Strings(names)
		sum := sha256.Sum256([]byte(os.Getenv("GREETING_SECRET")))
		_ = json.NewEncoder(w).Encode(map[string]any{"set": names, "greeting_digest": hex.EncodeToString(sum[:])[:8], "version": version})
	})
	http.HandleFunc("/db", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
		if err != nil {
			http.Error(w, "connect: "+err.Error(), 500)
			return
		}
		defer conn.Close(ctx)
		var one int
		if err := conn.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil {
			http.Error(w, "query: "+err.Error(), 500)
			return
		}
		if _, err := conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS visits (at timestamptz default now())"); err != nil {
			http.Error(w, "ddl: "+err.Error(), 500)
			return
		}
		fmt.Fprint(w, "ok")
	})
	log.Printf("hello %s listening on :%s", version, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
