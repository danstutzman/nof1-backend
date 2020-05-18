package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"bytes"
	"html/template"
	"net/http"
	"time"
)

var t = template.Must(template.New("recordings").Funcs(template.FuncMap{
	"formatTimestamp": func(secondsSinceEpoch float64) string {
		return time.Unix(int64(secondsSinceEpoch), 0).Format("2006-01-02 15:04:05 UTC")
	},
}).Parse(`<html>
	<head>
		<style>
			body { font-family: sans-serif; }
			th { text-align: left; }
		</style>
	</head>
	<body>
		<h1>Recordings</h1>
		<table>
			<tr>
				<th>Id</th>
				<th>UserId</th>
				<th>IdOnClient</th>
				<th>RecordedAtOnClient</th>
				<th>UploadedAt</th>
				<th>Filename</th>
				<th>MimeType</th>
				<th>Size</th>
				<th>MetadataJson</th>
			</tr>
			{{range .Recordings}}
				<tr>
					<td>{{.Id}}</td>
					<td>{{.UserId}}</td>
					<td>{{.IdOnClient}}</td>
					<td>{{formatTimestamp .RecordedAtOnClient}}</td>
					<td>{{.UploadedAt}}</td>
					<td>{{.Filename}}</td>
					<td>{{.MimeType}}</td>
					<td>{{.Size}}</td>
					<td>{{.MetadataJson}}</td>
				</tr>
			{{end}}
		</table>
	</body>
</html>`))

func (webapp *WebApp) getRecordings(r *http.Request,
	browser *db.BrowsersRow) Response {

	data := struct {
		Recordings []db.RecordingsRow
	}{
		Recordings: webapp.model.GetRecordings(),
	}

	var content bytes.Buffer
	err := t.Execute(&content, data)
	if err != nil {
		panic(err)
	}

	return BytesResponse{
		content:     content.Bytes(),
		contentType: "text/html; charset=utf-8",
	}
}
