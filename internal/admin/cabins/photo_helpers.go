package cabins

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"
	"sync"
	"travelando/internal/system/services"
	"travelando/internal/system/storage"

	"golang.org/x/sync/errgroup"
)

const maxPhotoWorkers = 4

type errCollector struct {
	mu   sync.Mutex
	errs []error
}

func (c *errCollector) add(err error) {
	if err == nil {
		return
	}
	c.mu.Lock()
	c.errs = append(c.errs, err)
	c.mu.Unlock()
}

func (c *errCollector) err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return errors.Join(c.errs...)
}

// DeleteCabinPhotos borra primero el registro en BD y, solo si salió bien,
// el archivo. Así nunca queda una fila apuntando a un archivo inexistente.
func DeleteCabinPhotos(ctx context.Context, svc services.CabinService, cabinID int64, photoURLs []string) error {
	st := services.ServiceSetFromCtx(ctx).Storage
	if st == nil {
		return errors.New("storage service unavailable")
	}

	var (
		g    errgroup.Group
		errs errCollector
	)
	g.SetLimit(maxPhotoWorkers)

	for _, u := range photoURLs {
		g.Go(func() error {
			rel := strings.TrimPrefix(u, "/uploads/")
			if !safeRelPath(rel) {
				errs.add(errors.New("invalid photo path: " + u))
				return nil
			}
			if err := svc.DeleteCabinPhoto(cabinID, u); err != nil {
				errs.add(err)
				return nil
			}
			errs.add(st.DeletePhoto(rel))
			return nil
		})
	}
	_ = g.Wait()
	return errs.err()
}

// SaveCabinPhotos sube los archivos con concurrencia limitada. Si falla el
// registro en BD, elimina el archivo recién guardado para no dejar huérfanos.
func SaveCabinPhotos(ctx context.Context, svc services.CabinService, cabinID int64, files []*multipart.FileHeader) error {
	st := services.ServiceSetFromCtx(ctx).Storage
	if st == nil {
		return errors.New("storage service unavailable")
	}

	var (
		g    errgroup.Group
		errs errCollector
	)
	g.SetLimit(maxPhotoWorkers)

	for _, fh := range files {
		g.Go(func() error {
			f, err := fh.Open()
			if err != nil {
				errs.add(err)
				return nil
			}
			rel, err := st.SavePhoto(f, fh)
			f.Close()
			if err != nil {
				errs.add(err)
				return nil
			}
			if err := svc.AddCabinPhoto(cabinID, storage.PublicURL(rel)); err != nil {
				errs.add(err)
				errs.add(st.DeletePhoto(rel))
			}
			return nil
		})
	}
	_ = g.Wait()
	return errs.err()
}

// OwnedPhotos devuelve solo las URLs solicitadas que realmente pertenecen a la cabaña.
func OwnedPhotos(existing, requested []string) []string {
	own := make(map[string]struct{}, len(existing))
	for _, u := range existing {
		own[u] = struct{}{}
	}
	out := make([]string, 0, len(requested))
	for _, u := range requested {
		if _, ok := own[u]; ok {
			out = append(out, u)
		}
	}
	return out
}

func safeRelPath(p string) bool {
	return p != "" && !strings.Contains(p, "..") && !strings.HasPrefix(p, "/")
}
