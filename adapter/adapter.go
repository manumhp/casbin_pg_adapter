package adapter

import (
	casbinmodel "casbin_adapter/casbinModel"
	"errors"
	"fmt"
	"strings"

	"database/sql"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	_ "github.com/lib/pq" //postgres driver : https://stackoverflow.com/questions/52789531/how-do-i-solve-panic-sql-unknown-driver-postgres-forgotten-import/52791919
)

// type CasbinRule struct {
// 	PType, V0, V1, V2, V3, V4, V5 string
// }

//Adapter represents the adapter for policy storage
type Adapter struct {
	db         *sql.DB
	tableNames []string
	filtered   bool
}

//DSN converts sql string to format for connecting to db
// func DSN(data PostgresqlInfo) string {

// 	t := "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable"
// 	connectionString := fmt.Sprintf(t, data.Host, data.Port, data.Username, data.Password, data.Dbname)
// 	return connectionString
// }

//NewAdapter creates the adapter object from a postgres url string
func NewAdapter(arg interface{}) (*Adapter, error) {
	if connURL, ok := arg.(string); ok {
		// SQL, err := sql.Open("postgres", DSN(connURL))
		SQL, err := sql.Open("postgres", connURL)
		if err != nil {
			return nil, err
		}
		a := &Adapter{db: SQL}

		if err := a.createTablesifNotExisting(); err != nil {
			return nil, fmt.Errorf("")
		}
		return a, nil
	}
	return nil, fmt.Errorf("Please pass in a postgresql url ")
}

func NewAdapterByDB(db *sql.DB) (*Adapter, error) {
	a := &Adapter{db: db}

	if err := a.createTablesifNotExisting(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Adapter) createTablesifNotExisting() error {
	// Roles table
	_, err := a.db.Exec("CREATE table IF NOT EXISTS t_role_policy_mapping (role_name VARCHAR(100), rights VARCHAR(100), PRIMARY KEY(role_name));")
	if err != nil {
		return err
		panic(err)
	}

	// Role-user mapping table
	_, err = a.db.Exec("CREATE table IF NOT EXISTS t_user_role_mapping (user_id VARCHAR(100), project_id VARCHAR(100), role_name VARCHAR(100), data_insertion_ts DATETIME, PRIMARY KEY(user_id, project_id));")
	if err != nil {
		return err
		panic(err)
	}
	return nil
}

func (a *Adapter) Close() error {
	if a != nil && a.db != nil {
		return a.db.Close()
	}
	return nil
}

func (a *Adapter) LoadPolicy(model model.Model) error {

	var rolePolicyRules = casbinmodel.CasbinRules{}
	var userRowRules = casbinmodel.CasbinRules{}

	query_role_policy_rows :=
		`select 
		distinct 'p' as PType ,concat(turm.project_id, '_', trpm.rolename) as v0,
		turm.project_id as v1, trpm.rights as v2
	from 
		public.t_role_policy_mapping trpm, public.t_user_role_mapping turm where 
		trpm.rolename = turm .role_name;`

	role_policy_rows, err := a.db.Query(query_role_policy_rows)

	if err != nil {
		return err
	}

	query_user_role_rows :=
		`select 
		distinct 'g' as PTYpe, turm.user_id as v0, concat(turm.project_id, '_', trpm.rolename) as v1 
	from 
		public.t_role_policy_mapping trpm inner join public.t_user_role_mapping turm 
		on trpm.rolename = turm .role_name;`

	user_role_rows, err := a.db.Query(query_user_role_rows)
	if err != nil {
		return err
	}

	defer role_policy_rows.Close()

	defer user_role_rows.Close()

	for role_policy_rows.Next() {
		err = role_policy_rows.Scan(&rolePolicyRules.PType, &rolePolicyRules.V0, &rolePolicyRules.V1, &rolePolicyRules.V2)
		persist.LoadPolicyLine(rolePolicyRules.String(), model)
	}

	for user_role_rows.Next() {
		err = user_role_rows.Scan(&userRowRules.PType, &userRowRules.V0, &userRowRules.V1)
		persist.LoadPolicyLine(userRowRules.String(), model)
	}

	a.filtered = false

	return nil

}

// func (a *Adapter) LoadPolicyLine() {} Already implemented in persist

//AddPolicy adds a rule to a user (g format: g, user, project, role)
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	// line := savePolicyLine(ptype, rule)
	projectName := strings.Split(rule[1], ".")[0]
	roleName := strings.Split(rule[1], ".")[1]
	insertQuery := `INSERT into public.t_user_role_mapping (user_id, project_id, role_name) VALUES($1, $2, $3) ON CONFLICT DO NOTHING;`
	// insertQuery := `UPSERT into public.t_user_role_mapping (user_id, project_id, role_name) VALUES($1, $2, $3) ON CONFLICT DO NOTHING;`
	_, error := a.db.Exec(insertQuery, rule[0], projectName, roleName)
	if error != nil {
		return error
	}
	return nil
}

func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return errors.New("not implemented")
}

func (a *Adapter) RemovePolicy(sec string, ptype string, policy []string) error {
	return errors.New("not implemented")
}

// SavePolicy saves policy to database.
func (a *Adapter) SavePolicy(model model.Model) error {
	return errors.New("not implemented")
	// tx, err := a.db.Begin()
	// if err != nil {
	// 	return fmt.Errorf("start DB transaction: %v", err)
	// }
	// defer tx.Close()

	// _, err = tx.Model((*CasbinRule)(nil)).Where("id IS NOT NULL").Delete()
	// if err != nil {
	// 	return err
	// }

	// var lines []*CasbinRule

	// for ptype, ast := range model["p"] {
	// 	for _, rule := range ast.Policy {
	// 		line := savePolicyLine(ptype, rule)
	// 		lines = append(lines, line)
	// 	}
	// }

	// for ptype, ast := range model["g"] {
	// 	for _, rule := range ast.Policy {
	// 		line := savePolicyLine(ptype, rule)
	// 		lines = append(lines, line)
	// 	}
	// }

	// if len(lines) > 0 {
	// 	_, err = tx.Model(&lines).
	// 		OnConflict("DO NOTHING").
	// 		Insert()
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// err = tx.Commit()
	// if err != nil {
	// 	return fmt.Errorf("commit DB transaction: %v", err)
	// }

	// return nil
}
