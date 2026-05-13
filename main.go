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

        log.Println("服务器已启动: http://localhost:8089")
        if err := http.ListenAndServe(":8089", nil); err != nil {
                log.Fatal(err)
        }
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./tracker.db") // 或 sqlite3
	if err != nil {
		log.Fatal(err)
	}

	// 1. 创建资产表
	queryItems := `CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		expires_at DATETIME NOT NULL
	);`
	db.Exec(queryItems)

	// 2. 创建用户表
	queryUsers := `CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password TEXT NOT NULL
	);`
	db.Exec(queryUsers)

	// 3. 如果用户表是空的，初始化注入默认账号密码 admin / admin
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count == 0 {
		db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", "admin", "admin")
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
		username := r.FormValue("username")
		password := r.FormValue("password")

		var dbPassword string
		// 去数据库查找该用户的密码
		err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&dbPassword)
		
		// 查到了用户且密码完全一致
		if err == nil && password == dbPassword {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "authenticated", Path: "/"})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		} else {
			data := map[string]interface{}{"HasError": true}
			tmpl.ExecuteTemplate(w, "login.html", data)
			return
		}
	}

	tmpl.ExecuteTemplate(w, "login.html", map[string]interface{}{"HasError": false})
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

		var dbPassword string
		// 验证当前原密码是否正确
		err := db.QueryRow("SELECT password FROM users WHERE username = 'admin'").Scan(&dbPassword)
		
		if err == nil && oldPwd == dbPassword && newPwd != "" {
			// 核心：直接更新数据库文件
			_, updateErr := db.Exec("UPDATE users SET password = ? WHERE username = 'admin'", newPwd)
			if updateErr == nil {
				// 修改成功后强制安全注销，让其用新密码重新登录
				http.Redirect(w, r, "/logout", http.StatusSeeOther)
				return
			}
		}
	}
	// 如果失败，直接弹回首页
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
