package sqldb

import (
	"encoding/json"

	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
)

const WorkflowHistoryTableName = "argo_workflow_history"

type WorkflowHistoryRepository interface {
	AddWorkflowHistory(wf *wfv1.Workflow) error
	ListWorkflowHistory(namespace string, limit, offset int) ([]wfv1.Workflow, error)
	GetWorkflowHistory(namespace string, uid string) (*wfv1.Workflow, error)
	DeleteWorkflowHistory(namespace string, uid string) error
}

type workflowHistoryRepository struct {
	sqlbuilder.Database
}

func NewWorkflowHistoryRepository(database sqlbuilder.Database) WorkflowHistoryRepository {
	return &workflowHistoryRepository{Database: database}
}

func (r *workflowHistoryRepository) AddWorkflowHistory(wf *wfv1.Workflow) error {
	err := r.DeleteWorkflowHistory(wf.Namespace, string(wf.UID))
	if err != nil {
		return err
	}
	wfDB, err := convert(wf)
	if err != nil {
		return err
	}
	_, err = r.Collection(WorkflowHistoryTableName).Insert(wfDB)
	return err
}

func (r *workflowHistoryRepository) ListWorkflowHistory(namespace string, limit int, offset int) ([]wfv1.Workflow, error) {
	var wfDBs []WorkflowDB
	err := r.Collection(WorkflowHistoryTableName).
		Find().
		Where(namespaceEqual(namespace)).
		OrderBy("-startedat").
		Limit(limit).
		Offset(offset).
		All(&wfDBs)
	return wfDB2wf(wfDBs), err
}

func namespaceEqual(namespace string) db.Cond {
	if namespace == "" {
		return db.Cond{}
	} else {
		return db.Cond{"namespace": namespace}
	}
}

func (r *workflowHistoryRepository) GetWorkflowHistory(namespace string, uid string) (*wfv1.Workflow, error) {
	rs := r.Collection(WorkflowHistoryTableName).
		Find().
		Where(db.Cond{"id": uid}).
		And(db.Cond{"namespace": namespace})
	exists, err := rs.Exists()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	wfDB := &WorkflowDB{}
	err = rs.One(wfDB)
	if err != nil {
		return nil, err
	}
	wf := &wfv1.Workflow{}
	err = json.Unmarshal([]byte(wfDB.Workflow), wf)
	if err != nil {
		return nil, err
	}
	return wf, nil
}

func (r *workflowHistoryRepository) DeleteWorkflowHistory(namespace string, uid string) error {
	return r.Collection(WorkflowHistoryTableName).
		Find().
		Where(db.Cond{"id": uid}).
		And(db.Cond{"namespace": namespace}).
		Delete()
}
