package main

import (
    "flag"
    "log"
    "net/http"
    "strings"
)

var f_wwwroot = flag.String("wwwroot", "www", "The path, relative to the cwd, where the static files live -- an empty string will disable static file serving")
var f_apiurl  = flag.String("apiurl", "/api", "The URL under which the QuickBump API is served")
var f_words   = flag.String("words", "", "The path of a word dictionary for generating relatively easy to remember question ids")
var f_qrurl   = flag.String("qrurl", "/qr", "The URL under which QR Codes are served")
var f_addr    = flag.String("addr", "localhost:8888", "The address and port to bind to when listening for HTTP connections. You leave the host blank, eg. :8888 should work")

// If a build includes the qrcode stuff, this should get set before
// main is executed
var QRModule http.Handler = nil

func main() {
    dblist := make([]string, len(DbRegistry))
    i := 0
    for name, _ := range DbRegistry {
        dblist[i] = name
        i++
    }
    f_db_default := ""
    if len(dblist) == 1 {
        f_db_default = dblist[0]
    }
    f_db := flag.String("db", f_db_default, "Database to use, choices are: " + strings.Join(dblist, ", "))

    flag.Parse()

    if *f_db == "" {
        log.Fatal("-db flag is required, run with -h for help")
    }

    wwwroot := *f_wwwroot
    apiurl  := *f_apiurl
    wordpath := *f_words

    fact, exists := DbRegistry[*f_db]
    if !exists {
        log.Fatal(*f_db + " not a valid database choice")
    }
    db := fact()
    defer db.Close()
    db.Reset()

    // Wordlist for question names
    qbhandler, err := NewQuickBumpHandler(db, wordpath)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("QuickBumpHandler loaded %d words", len(qbhandler.Wordlist))

    // QR Code handler
    if QRModule != nil {
        qrpath := *f_qrurl + "/"
        http.Handle(qrpath, http.StripPrefix(qrpath, QRModule))
        log.Print("QR Code module is loaded! Serving under ", qrpath)
    }

    // Serve API
    http.Handle(apiurl+"/", http.StripPrefix(apiurl, qbhandler))
    log.Print("QuickBumpHandler serving API under ",  apiurl)

    // Serve static
    if wwwroot != "" {
        log.Print("QuickBumpHandler serving static at / from ", wwwroot)
        http.Handle("/", http.FileServer(http.Dir(wwwroot)))
    }

    log.Print("Listening on ", *f_addr)
    if err := http.ListenAndServe(*f_addr, nil); err != nil {
        log.Fatal(err)
    }
}
