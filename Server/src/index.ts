import express from "express"
import Database from "better-sqlite3"
import path from "path"
import { fileURLToPath } from "url"

// CONSTANTS
const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const PORT = 8888
const DB_PATH = path.resolve(__dirname, "../../DB/test.db")

// INITIALIZE DATABASE
const db = new Database(DB_PATH)
db.pragma("journal_mode = WAL")

db.exec(`
  CREATE TABLE IF NOT EXISTS test (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message TEXT NOT NULL
  );
`)

// INITIALIZE SERVER
const app = express()
app.use(express.json())
app.get("/health", (req, res) => {
	res.json({ status: "ok", message: "Umbra Server is healthy." })
})
app.post("/test", (req, res) => {
	const message = req.body.message || "Hello from Umbra!"
	const stmt = db.prepare("INSERT INTO test (message) VALUES (?)")
	const result = stmt.run(message)

	res.json({ inserted: result.lastInsertRowid, message })
})
app.listen(PORT, () => {
	console.log(`✅ Umbra Server running at http://localhost:${PORT}`);
})