package pas_client

type Client struct {
	User_id    string `db:"user_id"`
	Login      string `db:"login"`
	Created_at string `db:"created_at"`
	Now_conns  int
}
