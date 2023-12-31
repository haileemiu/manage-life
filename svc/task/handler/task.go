package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/haileemiu/manage-life/ent"
	"github.com/haileemiu/manage-life/ent/task"
	"github.com/haileemiu/manage-life/pkg/res"
	"github.com/haileemiu/manage-life/svc/task/model"
)

type Task struct {
	ent *ent.Client
}

type DeleteResponse struct {
	Message string `json:"message"`
}

func New(entClient *ent.Client) *Task {
	return &Task{ent: entClient}
}

func (t Task) Routes(r chi.Router) {
	r.Get("/", t.list)
	r.Post("/", t.create)
	r.Get("/{id}", t.getByID)
	r.Put("/{id}", t.update)
	r.Delete("/{id}", t.delete)
}

func (t Task) list(w http.ResponseWriter, r *http.Request) {
	defaultPage := 1
	defaultPageSize := 10

	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("pageSize")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt <= 0 {
			pageInt = defaultPage
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeInt <= 0 {
			pageSizeInt = defaultPageSize
	}

	offset := (pageInt - 1) * pageSizeInt

	// TODO: return page & pagesize (specifically for defaults). metadata property
	tasks, err := t.ent.Task.Query().
			Limit(pageSizeInt).
			Offset(offset).
			Order(ent.Asc("created_at")).
			All(r.Context())

	if err != nil {
			http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
			return
	}

	var taskList []model.TaskItemResponse
	for _, entModel := range tasks {
		task := model.TaskItemResponse{
			ID: 						entModel.ID,
			Title:          entModel.Title,
			Notes:          entModel.Notes,
			IsTimeSenstive: entModel.IsTimeSenstive,
			IsImportant:    entModel.IsImportant,
			RemindAt:       entModel.RemindAt,
			DueAt:          entModel.DueAt,
		}
		taskList = append(taskList, task)
	}

	if err := json.NewEncoder(w).Encode(taskList); err != nil {
		log.Printf("Error occurred while encoding: %v", err)
		return
	}
}

func (t Task) create(w http.ResponseWriter, r *http.Request) {
	req := model.TaskCreateRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if ok, errs := req.Validate(); !ok {
		res.NewValidationErrorResponse(errs).Send(w)
		return
	}

	entTask, err := t.ent.Task.Create().
		SetTitle(req.Title).
		SetNotes(req.Notes).
		SetIsTimeSenstive(req.IsTimeSenstive).
		SetIsImportant(req.IsImportant).
		Save(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	task := model.TaskItemResponse{
		ID:             entTask.ID,
		Title:          entTask.Title,
		Notes:          entTask.Notes,
		IsTimeSenstive: entTask.IsTimeSenstive,
		IsImportant:    entTask.IsImportant,
		RemindAt:       entTask.RemindAt,
		DueAt:          entTask.DueAt,
	}

	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("Error occurred while encoding: %v", err)
		return
	}
}

func (t Task) getByID(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	task, err := t.ent.Task.Query().Where(task.ID(taskID)).Only(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	taskItem := model.TaskItemResponse{
		ID:             task.ID,
		Title:          task.Title,
		Notes:          task.Notes,
		IsTimeSenstive: task.IsTimeSenstive,
		IsImportant:    task.IsImportant,
		RemindAt:       task.RemindAt,
		DueAt:          task.DueAt,
	}

	if err := json.NewEncoder(w).Encode(taskItem); err != nil {
		log.Printf("Error occurred while encoding: %v", err)
		return
	}
}

func (t Task) update(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	req := model.TaskCreateRequest{}

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error occurred while decoding: %v", err)
		return
	}

	// TODO: a different validate for PUT? 
	// if ok, errs := req.Validate(); !ok {
	// 	res.NewValidationErrorResponse(errs).Send(w)
	// 	return
	// }

	entTask, err := t.ent.Task.UpdateOneID(taskID).
		SetTitle(req.Title).
		SetNotes(req.Notes).
		SetIsImportant(req.IsImportant).
		SetIsTimeSenstive(req.IsImportant).
		SetDueAt(req.DueAt).
		SetRemindAt(req.RemindAt).
		Save(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	task := model.TaskItemResponse{
		Title:          entTask.Title,
		Notes:          entTask.Notes,
		IsTimeSenstive: entTask.IsTimeSenstive,
		IsImportant:    entTask.IsImportant,
		RemindAt:       entTask.RemindAt,
		DueAt:          entTask.DueAt,
	}

	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("Error occurred while encoding: %v", err)
		return
	}
}

func (t Task) delete(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest) // Case: user sends non-integer
		return
	}

	err = t.ent.Task.DeleteOneID(taskID).Exec(r.Context())// Case: Item exists & successfully deleted

	if err != nil {
			if ent.IsNotFound(err) { // Case: Item does not exist
					w.WriteHeader(http.StatusNoContent)
					return
			}
			http.Error(w, "Failed to delete task", http.StatusInternalServerError)
			return
	}

	response := DeleteResponse{
			Message: "Task deleted successfully",
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
	}
}
