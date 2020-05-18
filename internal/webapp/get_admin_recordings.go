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
		return time.Unix(int64(secondsSinceEpoch), 0).Format(
			"2006-01-02 15:04:05 UTC")
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

		<form method='POST' action='/admin/recordings'>
			<table>
				<tr>
					<th>Id</th>
					<th>UserId</th>
					<th>IdOnClient</th>
					<th>UploadedAt</th>
					<th>Play</th>
					<th>Transcript</th>
				</tr>
				{{range .Recordings}}
					<tr>
						<td>{{.Id}}</td>
						<td>{{.UserId}}</td>
						<td>{{.IdOnClient}}</td>
						<td>{{.UploadedAt}}</td>
						<td>
							<audio controls>
								<source src='/admin/recordings/{{.UserId}}//{{.Filename}}'
									type='{{.MimeType}}' />
							</audio>
						</td>
						<td>
							<input name='{{.Id}}.transcript' value='{{.Transcript}}' />
						</td>
					</tr>
				{{end}}
			</table>

			<input type='submit' value='Save' />
		</form>
	</body>
</html>`))

func (webapp *WebApp) getAdminRecordings(r *http.Request,
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
