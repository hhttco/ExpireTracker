package main

import (
        "database/sql"
        "html/template"
        "log"
        "net/http"
        "strconv"
        "time"

        _ "github.com/mattn/go-sqlite3"
)

type Item struct {
        ID        int
        Name      string
        Type      string
        ExpiresAt time.Time
}

type ItemView struct {
        Item
        DaysLeft int
        Status   string // "danger", "warning", "safe", "expired"
}

var (
        db       *sql.DB
        tmpl     *template.Template
        // 系统管理员密码（内存持久，重启复位，可根据需要改为存入数据库）
        adminPassword = "admin" 
)

func main() {
        initDB()
        defer db.Close()

        tmpl = template.Must(template.ParseGlob("templates/*.html"))

        // 基础路由
        http.HandleFunc("/login", loginHandler)
        http.HandleFunc("/logout", logoutHandler)

        // 必须登录的路由
        http.HandleFunc("/", authMiddleware(indexHandler))
        http.HandleFunc("/add", authMiddleware(addHandler))
        http.HandleFunc("/edit", authMiddleware(editHandler))
        http.HandleFunc("/delete", authMiddleware(deleteHandler))
        http.HandleFunc("/change-password", authMiddleware(passwordHandler))

        log.Println("服务器已启动: http://localhost:8080")
        if err := http.ListenAndServe(":8080", nil); err != nil {
                log.Fatal(err)
        }
}

func initDB() {
        var err error
        db, err = sql.Open("sqlite3", "./tracker.db")
        if err != nil {
                log.Fatal(db)
        }

        query := `
        CREATE TABLE IF NOT EXISTS items (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                type TEXT NOT NULL,
                expires_at DATETIME NOT NULL
        );`
        if _, err = db.Exec(query); err != nil {
                log.Fatal(err)
        }
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                cookie, err := r.Cookie("session")
                if err != nil || cookie.Value != "authenticated" {
                        http.Redirect(w, r, "/login", http.StatusSeeOther)
                        return
                }
                next(w, r)
        }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
        rows, err := db.Query("SELECT id, name, type, expires_at FROM items ORDER BY expires_at ASC")
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        defer rows.Close()

        var viewItems []ItemView
        now := time.Now()

        for rows.Next() {
                var item Item
                var expiresAtStr string
                if err := rows.Scan(&item.ID, &item.Name, &item.Type, &expiresAtStr); err != nil {
                        continue
                }

                item.ExpiresAt, _ = time.Parse("2006-01-02T15:04:05Z", expiresAtStr)
                if item.ExpiresAt.IsZero() {
                        item.ExpiresAt, _ = time.Parse("2006-01-02 15:04:05", expiresAtStr)
                }

                // 计算精确天数差距
                daysLeft := int(item.ExpiresAt.Sub(now).Hours() / 24)
                status := "safe"

                if daysLeft < 0 {
                        status = "expired"
                } else if daysLeft <= 7 {
                        status = "danger"
                } else if daysLeft <= 30 {
                        status = "warning"
                }

                viewItems = append(viewItems, ItemView{
                        Item:     item,
                        DaysLeft: daysLeft,
                        Status:   status,
                })
        }

        // 传递当前日期给前端，用于日历默认值
        data := map[string]interface{}{
                "Items":       viewItems,
                "CurrentDate": now.Format("2006-01-02"),
        }

        tmpl.ExecuteTemplate(w, "index.html", data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
                if r.FormValue("username") == "admin" && r.FormValue("password") == adminPassword {
                        http.SetCookie(w, &http.Cookie{Name: "session", Value: "authenticated", Path: "/"})
                        http.Redirect(w, r, "/", http.StatusSeeOther)
                        return
                }
        }
        tmpl.ExecuteTemplate(w, "login.html", nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
        // 清除 Cookie 退出登录
        http.SetCookie(w, &http.Cookie{
                Name:   "session",
                Value:  "",
                Path:   "/",
                MaxAge: -1,
        })
        http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
                name := r.FormValue("name")
                itemType := r.FormValue("type")
                expiryDate := r.FormValue("expiry_date") + " 23:59:59"

                db.Exec("INSERT INTO items (name, type, expires_at) VALUES (?, ?, ?)", name, itemType, expiryDate)
        }
        http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
                id, _ := strconv.Atoi(r.FormValue("id"))
                expiryDate := r.FormValue("expiry_date") + " 23:59:59"

                db.Exec("UPDATE items SET expires_at = ? WHERE id = ?", expiryDate, id)
        }
        http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
                id, _ := strconv.Atoi(r.FormValue("id"))
                db.Exec("DELETE FROM items WHERE id = ?", id)
        }
        http.Redirect(w, r, "/", http.StatusSeeOther)
}

func passwordHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost {
                oldPwd := r.FormValue("old_password")
                newPwd := r.FormValue("new_password")

                if oldPwd == adminPassword && newPwd != "" {
                        adminPassword = newPwd
                        // 修改成功后强制重新登录
                        http.Redirect(w, r, "/logout", http.StatusSeeOther)
                        return
                }
        }
        http.Redirect(w, r, "/", http.StatusSeeOther)
}
