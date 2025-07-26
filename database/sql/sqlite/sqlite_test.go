// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sqlite

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/stretchr/testify/assert"
)

func TestSingleDdatabase(t *testing.T) {
	var version string
	_ = os.MkdirAll("../../../out", os.ModePerm)
	db, err := sql.Open("sqlite3", "file:../../../out/demo.db")
	assert.NoError(t, err, "failed to open db")
	defer func() { _ = db.Close() }()
	err = db.QueryRow(`SELECT sqlite_version()`).Scan(&version)
	assert.NoError(t, err, "failed to get sqlite version")
	t.Logf("sqlite version: %s", version)

	// 测试前清理已存在的表
	_, err = db.Exec(`DROP TABLE IF EXISTS users`)
	assert.NoError(t, err, "failed to drop users table")

	// 创建用户表
	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, age INTEGER)`)
	assert.NoError(t, err, "failed to create table")

	// 插入几条数据
	_, err = db.Exec(`INSERT INTO users (name, age) VALUES (?, ?)`, "Alice", 30)
	assert.NoError(t, err, "failed to insert Alice")
	_, err = db.Exec(`INSERT INTO users (name, age) VALUES (?, ?)`, "Bob", 25)
	assert.NoError(t, err, "failed to insert Bob")
	_, err = db.Exec(`INSERT INTO users (name, age) VALUES (?, ?)`, "Charlie", 28)
	assert.NoError(t, err, "failed to insert Charlie")

	// 读取并测试
	rows, err := db.Query(`SELECT id, name, age FROM users ORDER BY id`)
	assert.NoError(t, err, "failed to query users")
	defer func() { _ = rows.Close() }()
	var count int
	var users []struct {
		id   int
		name string
		age  int
	}
	for rows.Next() {
		var u struct {
			id   int
			name string
			age  int
		}
		err := rows.Scan(&u.id, &u.name, &u.age)
		assert.NoError(t, err, "failed to scan user")
		t.Logf("user: id=%d, name=%s, age=%d", u.id, u.name, u.age)
		users = append(users, u)
		count++
	}
	assert.Equal(t, 3, count, "expected 3 users")
	assert.Equal(t, "Alice", users[0].name)
	assert.Equal(t, 30, users[0].age)
	assert.Equal(t, "Bob", users[1].name)
	assert.Equal(t, 25, users[1].age)
	assert.Equal(t, "Charlie", users[2].name)
	assert.Equal(t, 28, users[2].age)
}
