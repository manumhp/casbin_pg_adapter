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

//Adapter represents the adapter for policy storage
type Adapter struct {
	db         *sql.DB
	schemaName string
	tableNames []string
	filtered   bool
}

//NewAdapter creates the adapter object from a postgres url string
func NewAdapter(arg interface{}) (*Adapter, error) {
	if connURL, ok := arg.(string); ok {
		SQL, err := sql.Open("postgres", connURL)
		if err != nil {
			return nil, err
		}
		a := &Adapter{db: SQL, schemaName: "public"}

		if err := a.createTablesifNotExisting(); err != nil {
			return nil, fmt.Errorf("")
		}
		return a, nil
	}
	return nil, fmt.Errorf("Please pass in a postgresql url ")
}

//NewAdapterByDB creates a new casbin adapter with the db pointer supplied as an argument
func NewAdapterByDB(db *sql.DB, schemaName string) (*Adapter, error) {
	a := &Adapter{db: db, schemaName: schemaName}

	if err := a.createTablesifNotExisting(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Adapter) createTablesifNotExisting() error {
	// Roles table
	schemaName := a.schemaName
	_, err := a.db.Exec("CREATE table IF NOT EXISTS " + schemaName + ".t_role_policy_mapping (role_name VARCHAR(100), rights VARCHAR(100), PRIMARY KEY(role_name, rights));")
	if err != nil {
		return err
	}

	// Role-user mapping table
	_, err = a.db.Exec("CREATE table IF NOT EXISTS " + schemaName + ".t_user_role_mapping (user_id VARCHAR(100), project_id VARCHAR(100), role_name VARCHAR(100), data_insertion_ts DATETIME, PRIMARY KEY(user_id, project_id));")
	if err != nil {
		return err
	}
	return nil
}

//Close closes the db connection in the adapter
func (a *Adapter) Close() error {
	if a != nil && a.db != nil {
		return a.db.Close()
	}
	return nil
}

//LoadPolicy loads data from the database and generates the access policy according to casbin requirements
func (a *Adapter) LoadPolicy(model model.Model) error {

	var rolePolicyRules = casbinmodel.CasbinRules{}
	var userRowRules = casbinmodel.CasbinRules{}
	schemaName := a.schemaName

	queryRolePolicyRows :=
		`select 
		distinct 'p' as PType ,concat(turm.project_id, '_', trpm.role_name) as v0,
		turm.project_id as v1, trpm.rights as v2
	from ` + schemaName + `.t_role_policy_mapping trpm, public.t_user_role_mapping turm where 
		trpm.role_name = turm .role_name;`

	rolePolicyRows, err := a.db.Query(queryRolePolicyRows)

	if err != nil {
		return err
	}

	queryUserRoleRows :=
		`select 
		distinct 'g' as PTYpe, turm.user_id as v0, concat(turm.project_id, '_', trpm.role_name) as v1 
	from ` + schemaName + `.t_role_policy_mapping trpm inner join public.t_user_role_mapping turm 
		on trpm.role_name = turm .role_name;`

	userRoleRows, err := a.db.Query(queryUserRoleRows)
	if err != nil {
		return err
	}

	defer rolePolicyRows.Close()

	defer userRoleRows.Close()

	for rolePolicyRows.Next() {
		err = rolePolicyRows.Scan(&rolePolicyRules.PType, &rolePolicyRules.V0, &rolePolicyRules.V1, &rolePolicyRules.V2)
		persist.LoadPolicyLine(rolePolicyRules.String(), model)
	}

	for userRoleRows.Next() {
		err = userRoleRows.Scan(&userRowRules.PType, &userRowRules.V0, &userRowRules.V1)
		persist.LoadPolicyLine(userRowRules.String(), model)
	}

	a.filtered = false

	return nil

}

// func (a *Adapter) LoadPolicyLine() {} Already implemented in persist

//AddPolicy adds a rule to a user (g format: g, user, project, role)
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	projectName := strings.Split(rule[1], ".")[0]
	roleName := strings.Split(rule[1], ".")[1]
	schemaName := a.schemaName
	insertQuery := `INSERT into ` + schemaName + `.t_user_role_mapping (user_id, project_id, role_name) VALUES($1, $2, $3) ON CONFLICT DO NOTHING;`
	// insertQuery := `INSERT into public.t_user_role_mapping (user_id, project_id, role_name) VALUES($1, $2, $3) ON CONFLICT UPDATE;`
	_, error := a.db.Exec(insertQuery, rule[0], projectName, roleName)
	if error != nil {
		return error
	}
	return nil
}

//RemoveFilteredPolicy removes a filtered policy from database
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return errors.New("not implemented")
}

//RemovePolicy removes a filtered policy from database
func (a *Adapter) RemovePolicy(sec string, ptype string, policy []string) error {
	return errors.New("not implemented")
}

// SavePolicy saves policy to database.
func (a *Adapter) SavePolicy(model model.Model) error {
	return errors.New("not implemented")
}
