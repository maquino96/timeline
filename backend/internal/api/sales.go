package api

import (
	"net/http"

	"github.com/maquino96/timeline/internal/models"
)

func (a *API) handleGetWatchItems(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.GetWatchItems(false)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (a *API) handleCreateWatchItem(w http.ResponseWriter, r *http.Request) {
	var item models.WatchItem
	if err := decodeJSON(r, &item); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if item.SearchTerm == "" {
		item.SearchTerm = item.Name
	}
	if err := a.store.AddWatchItem(&item); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	item.Active = true
	writeJSON(w, http.StatusCreated, item)
}

func (a *API) handleUpdateWatchItem(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var item models.WatchItem
	if err := decodeJSON(r, &item); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	item.ID = id
	if err := a.store.UpdateWatchItem(&item); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *API) handleDeleteWatchItem(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := a.store.DeleteWatchItem(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleGetSaleAlerts(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 25)
	offset := queryInt(r, "offset", 0)
	alerts, err := a.store.GetRecentSaleAlerts(14, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	total, _ := a.store.CountRecentSaleAlerts(14)
	w.Header().Set("X-Total-Count", itoaHeader(total))
	writeJSON(w, http.StatusOK, alerts)
}

func (a *API) handleDismissSaleAlert(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := a.store.DismissSaleAlert(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
