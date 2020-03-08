package db

import (
	"database/sql"
	"fmt"
	"gopkg.in/guregu/null.v3"
	"log"
	"time"
)

type CapabilitiesRow struct {
	Id                      int
	RequestId               int
	NavigatorAppCodeName    null.String
	NavigatorAppName        null.String
	NavigatorAppVersion     null.String
	NavigatorCookieEnabled  null.String
	NavigatorLanguage       null.String
	NavigatorLanguages      null.String
	NavigatorPlatform       null.String
	NavigatorOscpu          null.String
	NavigatorUserAgent      null.String
	NavigatorVendor         null.String
	NavigatorVendorSub      null.String
	ScreenWidth             null.String
	ScreenHeight            null.String
	WindowInnerWidth        null.String
	WindowInnerHeight       null.String
	DocBodyClientWidth      null.String
	DocBodyClientHeight     null.String
	DocElementClientWidth   null.String
	DocElementClientHeight  null.String
	WindowScreenAvailWidth  null.String
	WindowScreenAvailHeight null.String
	WindowDevicePixelRatio  null.String
	HasOnTouchStart         null.String
}

func assertCapabilitiesHasCorrectSchema(db *sql.DB) {
	query := `SELECT id, request_id, created_at,
		navigator_app_code_name,
		navigator_app_name,
		navigator_app_version,
		navigator_cookie_enabled,
		navigator_language,
		navigator_languages,
		navigator_platform,
		navigator_oscpu,
		navigator_user_agent,
		navigator_vendor,
		navigator_vendor_sub,
		screen_width,
		screen_height,
    window_inner_width,
    window_inner_height,
    doc_body_client_width,
    doc_body_client_height,
    doc_element_client_width,
    doc_element_client_height,
    window_screen_avail_width,
    window_screen_avail_height,
    window_device_pixel_ratio,
    has_on_touch_start
	  FROM capabilities LIMIT 1`
	if LOG {
		log.Println(query)
	}

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func InsertIntoCapabilities(db *sql.DB, row CapabilitiesRow) CapabilitiesRow {
	query := fmt.Sprintf(`INSERT INTO capabilities
	  (request_id, created_at,
		navigator_app_code_name,
		navigator_app_name,
		navigator_app_version,
		navigator_cookie_enabled,
		navigator_language,
		navigator_languages,
		navigator_platform,
		navigator_oscpu,
		navigator_user_agent,
		navigator_vendor,
		navigator_vendor_sub,
    screen_width,
    screen_height,
    window_inner_width,
    window_inner_height,
    doc_body_client_width,
    doc_body_client_height,
    doc_element_client_width,
    doc_element_client_height,
    window_screen_avail_width,
    window_screen_avail_height,
    window_device_pixel_ratio,
    has_on_touch_start)
    VALUES (%d, %s,
		  %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s,
		  %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`,
		row.RequestId,
		EscapeNanoTime(time.Now().UTC()),
		EscapeNullString(row.NavigatorAppCodeName),
		EscapeNullString(row.NavigatorAppName),
		EscapeNullString(row.NavigatorAppVersion),
		EscapeNullString(row.NavigatorCookieEnabled),
		EscapeNullString(row.NavigatorLanguage),
		EscapeNullString(row.NavigatorLanguages),
		EscapeNullString(row.NavigatorPlatform),
		EscapeNullString(row.NavigatorOscpu),
		EscapeNullString(row.NavigatorUserAgent),
		EscapeNullString(row.NavigatorVendor),
		EscapeNullString(row.NavigatorVendorSub),
		EscapeNullString(row.ScreenWidth),
		EscapeNullString(row.ScreenHeight),
		EscapeNullString(row.WindowInnerWidth),
		EscapeNullString(row.WindowInnerHeight),
		EscapeNullString(row.DocBodyClientWidth),
		EscapeNullString(row.DocBodyClientHeight),
		EscapeNullString(row.DocElementClientWidth),
		EscapeNullString(row.DocElementClientHeight),
		EscapeNullString(row.WindowScreenAvailWidth),
		EscapeNullString(row.WindowScreenAvailHeight),
		EscapeNullString(row.WindowDevicePixelRatio),
		EscapeNullString(row.HasOnTouchStart))
	if LOG {
		log.Println(query)
	}

	result, err := db.Exec(query)
	if err != nil {
		panic(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}
	row.Id = int(id)

	return row
}
