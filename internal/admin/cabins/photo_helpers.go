package cabins

import (
	"mime/multipart"
	"net/http"
	"strings"
	"sync"

	"travelando/internal/system/services"
	"travelando/internal/system/storage"
)

func DeleteCabinPhotos(w http.ResponseWriter, r *http.Request, svc services.CabinService, cabinID int64, photoURLs []string) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var inner sync.WaitGroup
		for _, photoURL := range photoURLs {
			inner.Add(1)
			go func(photoURL string) {
				defer inner.Done()
				_ = svc.DeleteCabinPhoto(cabinID, photoURL)
				_ = services.ServiceSetFromCtx(r.Context()).Storage.DeletePhoto(strings.TrimPrefix(photoURL, "/uploads/"))
			}(photoURL)
		}
		inner.Wait()
	}()
	wg.Wait()
}

func SaveCabinPhotos(w http.ResponseWriter, r *http.Request, svc services.CabinService, cabinID int64, files []*multipart.FileHeader) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		photoSvc := services.ServiceSetFromCtx(r.Context()).Storage
		var inner sync.WaitGroup
		for _, fh := range files {
			inner.Add(1)
			go func(fh *multipart.FileHeader) {
				defer inner.Done()
				f, err := fh.Open()
				if err != nil {
					return
				}
				rel, err := photoSvc.SavePhoto(f, fh)
				f.Close()
				if err != nil {
					return
				}
				_ = svc.AddCabinPhoto(cabinID, storage.PublicURL(rel))
			}(fh)
		}
		inner.Wait()
	}()
	wg.Wait()
}