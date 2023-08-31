package pg

import (
	"context"
	"github.com/jmoiron/sqlx"
)

func createTable(db *sqlx.DB) error {
	_, err := db.ExecContext(context.Background(),
		` CREATE TABLE IF NOT EXISTS users (
					 id serial PRIMARY KEY,
					 login varchar(450) NOT NULL UNIQUE,
					 password varchar(450) NOT NULL
				);

				CREATE TABLE IF NOT EXISTS orders (
					id serial PRIMARY KEY,
					number varchar(450) NOT NULL UNIQUE,
				    user_id INT,
					accrual bigint NOT NULL DEFAULT '0',
					withdrawn bigint NOT NULL DEFAULT '0',
					status varchar(450) NOT NULL,
					uploaded_at TIMESTAMP NOT NULL, --DEFAULT CURRENT_TIME,
				    CONSTRAINT fk_user
						FOREIGN KEY(user_id) 
						REFERENCES users(id)
				);
			`)

	return err
}
