package webapp

import (
	"bitbucket.org/danstutzman/nof1-backend/internal/db"
	"bytes"
	"html/template"
	"net/http"
	"time"
)

func formatTimestamp(secondsSinceEpoch float64) string {
	return time.Unix(int64(secondsSinceEpoch), 0).Format(
		"2006-01-02 15:04:05 UTC")
}

func formatToMountainTime(date time.Time) string {
	location, err := time.LoadLocation("America/Denver")
	if err != nil {
		panic(err)
	}
	return date.In(location).Format("15:04")
}

const TEMPLATE = `<html>
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
					<th>UploadedAt<br/>(Mountain Time)</th>
					<th>Play</th>
					<th>TranscriptAws</th>
					<th>TranscriptManual</th>
				</tr>
				{{range .Recordings}}
					<tr>
						<td>{{.Id}}</td>
						<td>{{.UserId}}</td>
						<td>{{.IdOnClient}}</td>
						<td>{{formatToMountainTime .UploadedAt}}</td>
						<td>
							<audio controls>
								<source src='/admin/recordings/{{.UserId}}//{{.Filename}}'
									type='{{.MimeType}}' />
							</audio>
						</td>
						<td>{{.TranscriptAws}}</td>
						<td>
							<input name='{{.Id}}.transcriptManual'
								value='{{.TranscriptManual}}' />
						</td>
					</tr>
				{{end}}
			</table>

			<input type='submit' value='Save' />
		</form>
	</body>
</html>`

var t = template.Must(template.New("recordings").Funcs(template.FuncMap{
	"formatTimestamp":      formatTimestamp,
	"formatToMountainTime": formatToMountainTime,
}).Parse(TEMPLATE))

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
