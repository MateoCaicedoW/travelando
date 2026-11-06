package components

import (
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func photoNodes(photos []models.CabinPhoto) []Node {
	nodes := make([]Node, 0, len(photos))
	for _, photo := range photos {
		nodes = append(nodes, Div(Class("relative"),
			A(Href(photo.URL), Target("_blank"), Rel("noopener"),
				Img(Src(photo.URL), Alt("Foto de cabaña"), Class("h-24 w-full rounded-lg object-cover")),
			),
			Button(Type("button"), Class("absolute right-1 top-1 rounded-full bg-black/60 px-2 py-1 text-xs text-white"), Text("×"),
				Attr("onclick", "const p=this.parentElement; const f=p.closest('form'); const i=document.createElement('input'); i.type='hidden'; i.name='removed_photos'; i.value='"+photo.URL+"'; f.appendChild(i); p.remove();"),
			),
		))
	}
	return nodes
}

func CabinPhotoManager(cabinID int64, photos []models.CabinPhoto, uploadURL, deleteURL string) Node {
	if cabinID == 0 {
		uploadURL = ""
	}
	return Div(ID("cabin-photo-manager"), Attr("data-controller", "photo-dropzone"), Attr("data-action", "dragover->photo-dropzone#dragover dragleave->photo-dropzone#dragleave drop->photo-dropzone#drop"), Class("space-y-4"),
		Div(Attr("data-photo-dropzone-target", "stack"), Class("grid grid-cols-2 gap-3 sm:grid-cols-4"),
			Group(photoNodes(photos)),
		),
		Div(Class("flex min-h-28 cursor-pointer flex-col items-center justify-center rounded-xl border-2 border-dashed border-ink/20 bg-cream/40 p-4 text-center transition hover:border-brand"),
			Input(Type("file"), Name("photos"), Attr("multiple", "multiple"), Attr("accept", "image/*"), Class("hidden"), Attr("data-photo-dropzone-target", "input")),
			P(Class("text-sm font-medium text-ink"), Text("Arrastra fotos aquí o haz clic para seleccionar")),
			P(Class("mt-1 text-xs text-ink/60"), Text("PNG, JPG o WEBP. Se pueden seleccionar varias.")),
			P(Class("photo-dropzone-error mt-1 text-xs text-red-600"), Text("")),
			Div(Attr("data-photo-dropzone-target", "preview"), Class("mt-3 grid w-full grid-cols-3 gap-2")),
		),
	)
}