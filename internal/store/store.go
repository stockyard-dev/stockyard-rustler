package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type ScrapeJob struct {
	ID string `json:"id"`
	Name string `json:"name"`
	URL string `json:"url"`
	Selector string `json:"selector"`
	Schedule string `json:"schedule"`
	LastResult string `json:"last_result"`
	Status string `json:"status"`
	RunCount int `json:"run_count"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"rustler.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS scrape_jobs(id TEXT PRIMARY KEY,name TEXT NOT NULL,url TEXT DEFAULT '',selector TEXT DEFAULT '',schedule TEXT DEFAULT '',last_result TEXT DEFAULT '',status TEXT DEFAULT 'active',run_count INTEGER DEFAULT 0,created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *ScrapeJob)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO scrape_jobs(id,name,url,selector,schedule,last_result,status,run_count,created_at)VALUES(?,?,?,?,?,?,?,?,?)`,e.ID,e.Name,e.URL,e.Selector,e.Schedule,e.LastResult,e.Status,e.RunCount,e.CreatedAt);return err}
func(d *DB)Get(id string)*ScrapeJob{var e ScrapeJob;if d.db.QueryRow(`SELECT id,name,url,selector,schedule,last_result,status,run_count,created_at FROM scrape_jobs WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.URL,&e.Selector,&e.Schedule,&e.LastResult,&e.Status,&e.RunCount,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]ScrapeJob{rows,_:=d.db.Query(`SELECT id,name,url,selector,schedule,last_result,status,run_count,created_at FROM scrape_jobs ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []ScrapeJob;for rows.Next(){var e ScrapeJob;rows.Scan(&e.ID,&e.Name,&e.URL,&e.Selector,&e.Schedule,&e.LastResult,&e.Status,&e.RunCount,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Update(e *ScrapeJob)error{_,err:=d.db.Exec(`UPDATE scrape_jobs SET name=?,url=?,selector=?,schedule=?,last_result=?,status=?,run_count=? WHERE id=?`,e.Name,e.URL,e.Selector,e.Schedule,e.LastResult,e.Status,e.RunCount,e.ID);return err}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM scrape_jobs WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM scrape_jobs`).Scan(&n);return n}

func(d *DB)Search(q string, filters map[string]string)[]ScrapeJob{
    where:="1=1"
    args:=[]any{}
    if q!=""{
        where+=" AND (name LIKE ?)"
        args=append(args,"%"+q+"%");
    }
    if v,ok:=filters["status"];ok&&v!=""{where+=" AND status=?";args=append(args,v)}
    rows,_:=d.db.Query(`SELECT id,name,url,selector,schedule,last_result,status,run_count,created_at FROM scrape_jobs WHERE `+where+` ORDER BY created_at DESC`,args...)
    if rows==nil{return nil};defer rows.Close()
    var o []ScrapeJob;for rows.Next(){var e ScrapeJob;rows.Scan(&e.ID,&e.Name,&e.URL,&e.Selector,&e.Schedule,&e.LastResult,&e.Status,&e.RunCount,&e.CreatedAt);o=append(o,e)};return o
}

func(d *DB)Stats()map[string]any{
    m:=map[string]any{"total":d.Count()}
    rows,_:=d.db.Query(`SELECT status,COUNT(*) FROM scrape_jobs GROUP BY status`)
    if rows!=nil{defer rows.Close();by:=map[string]int{};for rows.Next(){var s string;var c int;rows.Scan(&s,&c);by[s]=c};m["by_status"]=by}
    return m
}
