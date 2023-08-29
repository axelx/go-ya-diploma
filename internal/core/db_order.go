package core

import (
	"context"
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func FindOrder(db *sqlx.DB, lg *zap.Logger, orderID string) (models.Order, error) {
	row := db.QueryRowContext(context.Background(),
		`SELECT user_id, number FROM orders WHERE number = $1`, orderID)

	var o models.Order
	err := row.Scan(&o.UserID, &o.Number)
	if err != nil {
		lg.Info("FindOrder: order not found", zap.String("about ERR", err.Error()))
		return models.Order{}, err
	}
	return o, nil
}

func FindOrders(db *sqlx.DB, lg *zap.Logger, userID int) ([]models.Order, error) {

	rows, err := db.QueryContext(context.Background(),
		` SELECT number,accrual,withdrawn,status,uploaded_at FROM orders WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []models.Order{}

	for rows.Next() {
		var o models.Order
		err = rows.Scan(&o.Number, &o.Accrual, &o.Withdrawn, &o.Status, &o.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	err = rows.Err()
	if err != nil {
		lg.Info("Error FindOrders:", zap.String("about ERR", err.Error()))
		return nil, err
	}
	return orders, nil
}

func AddOrder(db *sqlx.DB, lg *zap.Logger, userID int, order string, withdrawn float64) error {

	if withdrawn > 0 {
		w := int(withdrawn * 100)
		_, err := db.ExecContext(context.Background(),
			`INSERT INTO orders (number, status, user_id, withdrawn, uploaded_at) VALUES ($1, $2, $3, $4, NOW())`,
			order, "NEW", userID, w)
		if err != nil {
			lg.Info("Error AddOrder:", zap.String("about ERR", err.Error()))
			return err
		}

	} else {
		_, err := db.ExecContext(context.Background(),
			`INSERT INTO orders (number, status, user_id, uploaded_at) VALUES ($1, $2, $3, NOW())`, order, "NEW", userID)
		if err != nil {
			lg.Info("Error AddOrder:", zap.String("about ERR", err.Error()))
			return err
		}
	}

	return nil
}

func UpdateStatusOrder(db *sqlx.DB, lg *zap.Logger, orderID, status string, accrual int) error {
	_, err := db.ExecContext(context.Background(),
		`UPDATE orders SET status = $1,  accrual = $2 WHERE number = $3`, status, accrual, orderID)
	//_, err := db.ExecContext(context.Background(),
	//	`UPDATE orders SET status = $1 WHERE number = $2`, status, orderID)
	if err != nil {
		lg.Info("Error AddOrder:", zap.String("about ERR", err.Error()))
		return err
	}
	return err
}
