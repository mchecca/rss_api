package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"mchecca.com/rss_api/models"
)

func writeJSON(w http.ResponseWriter, i interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(i)
}

func authenticate(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, _ := r.BasicAuth()
		auth := models.AuthenticateUser(username, password)
		if !auth {
			w.Header().Add("Www-Authenticate", "Basic realm=\"Login Required\"")
			w.WriteHeader(401)
			fmt.Fprintln(w, "Authentication Required")
		} else {
			h(w, r)
		}
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string][]string{"apiLevels": []string{"v1-2"}}
	writeJSON(w, response)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _ := r.BasicAuth()
	response := map[string]interface{}{
		"userId":             username,
		"displayName":        username,
		"lastLoginTimestamp": time.Now().Unix(),
		"avatar":             nil,
	}
	writeJSON(w, response)
}

func versionStatusHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"version": "6.0.5"}
	writeJSON(w, response)
}

func feedsHandler(w http.ResponseWriter, r *http.Request) {
	feeds := models.NCGetAllFeeds()
	response := map[string][]models.NCFeed{"feeds": feeds}
	writeJSON(w, response)
}

func foldersHandler(w http.ResponseWriter, r *http.Request) {
	folders := models.NCGetAllFolders()
	response := map[string][]models.NCFolder{"folders": folders}
	writeJSON(w, response)
}

func queryValueOrDefault(v url.Values, key string, defaultValue string) string {
	value := defaultValue
	_, exists := v[key]
	if exists {
		value = v.Get(key)
	}
	return value
}

func queryIntValueOrDefault(v url.Values, key string, defaultValue int) int {
	value := defaultValue
	_, exists := v[key]
	if exists {
		valueString := v.Get(key)
		valueInt, err := strconv.Atoi(valueString)
		if err == nil {
			value = valueInt
		}
	}
	return value
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	// batchSize := queryIntValueOrDefault(queryParams, "batchSize", -1)
	// offset := queryIntValueOrDefault(queryParams, "offset", -1)
	// getRead := queryValueOrDefault(queryParams, "getRead", "true") == "true"
	// oldestFirst := queryValueOrDefault(queryParams, "oldestFirst", "false")
	id := queryIntValueOrDefault(queryParams, "id", -1)
	typeQuery := queryIntValueOrDefault(queryParams, "type", -1)
	if id == -1 {
		w.WriteHeader(400)
		fmt.Fprint(w, "ID not specified")
		return
	}
	if typeQuery == -1 {
		w.WriteHeader(400)
		fmt.Fprint(w, "type not specified")
		return
	}
	var items []models.NCItem

	// items_query = models.Item.select()
	if typeQuery == 0 {
		panic("Unhandled type: " + strconv.Itoa(typeQuery))
		//     items_query = items_query.where(models.Item.feed_id == id_)
	} else if typeQuery == 1 {
		panic("Unhandled type: " + strconv.Itoa(typeQuery))
		//     feed_ids = [f.id for f in models.Feed.select(models.Feed.id).where(
		//         models.Feed.folder_id == id_)]
		//     items_query = items_query.where(models.Item.feed_id.in_(feed_ids))
	} else if typeQuery == 2 {
		panic("Unhandled type: " + strconv.Itoa(typeQuery))
		//     items_query = items_query.where(models.Item.starred)
	} else if typeQuery == 3 {
		items = models.NCQueryItems()
	} else {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Unknown type: %d\n", typeQuery)
		return
	}
	// items = [_item_to_json(i) for i in items_query]
	// if offset > 0:
	//     items_query = items_query.where(models.Item.id <= offset)
	// if not getRead:
	//     items_query = items_query.where(models.Item.read == False)
	// if oldestFirst:
	//     items_query = items_query.order_by(models.Item.pubDate)
	// else:
	//     items_query = items_query.order_by(models.Item.pubDate.desc())
	// if batchSize > 0:
	//     items_query = items_query.limit(batchSize)

	response := map[string][]models.NCItem{"items": items}
	writeJSON(w, response)
}

func main() {
	http.HandleFunc("/index.php/apps/news/api", authenticate(indexHandler))
	http.HandleFunc("/index.php/apps/news/api/v1-2/user", authenticate(userHandler))
	http.HandleFunc("/index.php/apps/news/api/v1-2/version", authenticate(versionStatusHandler))
	http.HandleFunc("/index.php/apps/news/api/v1-2/status", authenticate(versionStatusHandler))
	http.HandleFunc("/index.php/apps/news/api/v1-2/feeds", authenticate(feedsHandler))
	http.HandleFunc("/index.php/apps/news/api/v1-2/folders", authenticate(foldersHandler))
	http.HandleFunc("/index.php/apps/news/api/v1-2/items", authenticate(itemsHandler))
	log.Println("Listening on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

/*
@api.route('/items/<int:item_id>/read', methods=['PUT'])
def item_read(item_id):
    """Mark an item as read."""
    i = models.Item.get_by_id(item_id)
    i.read = True
    i.save()
    return json_response({})


@api.route('/items/<int:item_id>/unread', methods=['PUT'])
def item_unread(item_id):
    """Mark an item as unread."""
    i = models.Item.get_by_id(item_id)
    i.read = False
    i.save()
    return json_response({})


@api.route('/feeds/<int:feed_id>/read', methods=['PUT'])
def feed_read(feed_id):
    """Mark items of a feed as read."""
    newestItemId = int(flask.request.args['newestItemId'])
    if feed_id == 0:
        models.Item.update(read=True).where(models.Item.starred).execute()
    else:
        models.Item.update(read=True).where(
            (models.Item.feed_id == feed_id) & (models.Item.id <= newestItemId)).execute()
    return json_response({})


@api.route('/items/updated')
def items_updated():
    """Get updated items."""
    lastModifiedSeconds = int(flask.request.args['lastModified'])
    lastModified = datetime.datetime.fromtimestamp(lastModifiedSeconds)
    type_ = int(flask.request.args['type'])
    id_ = int(flask.request.args['id'])
    items_query = None
    if type_ == 0:
        items_query = models.Item.select().where(
            (models.Item.feed_id == id_) & (models.Item.updated >= lastModified)
        )
    elif type_ == 1:
        feed_ids = [f.id for f in models.Feed.select(models.Feed.id).where(
            (models.Feed.folder_id == 1) & (models.Item.updated >= lastModified)
        )]
        items_query = models.Item.select().where(
            (models.Item.feed_id.in_(feed_ids)) & (models.Item.updated >= lastModified)
        )
    elif type_ == 2:
        items_query = models.Item.select().where(
            (models.Item.starred) & (models.Item.updated >= lastModified)
        )
    elif type_ == 3:
        items_query = models.Item.select().where(models.Item.updated >= lastModified)
    else:
        logging.exception(Exception('Unknown type: {0}'.format(type_)))
        flask.abort(500)
    items = [_item_to_json(i) for i in items_query]
    return json_response({'items': items})


@api.route('/items/read/multiple', methods=['PUT'])
def items_read_multiple():
    """Mark multiple items as read."""
    items = flask.request.get_json().get('items', [])
    models.Item.update(read=True).where(models.Item.id.in_(items)).execute()
    return json_response({})


@api.route('/items/unread/multiple', methods=['PUT'])
def items_unread_multiple():
    """Mark multiple items as unread."""
    items = flask.request.get_json().get('items', [])
    models.Item.update(read=False).where(models.Item.id.in_(items)).execute()
    return json_response({})


@api.route('/items/star/multiple', methods=['PUT'])
def items_star_multiple():
    """Mark multiple items as starred."""
    items = flask.request.get_json().get('items', [])
    for i in items:
        feed_id = int(i['feedId'])
        guidHash = i['guidHash']
        guid = base64.decodebytes(guidHash.encode()).decode()
        models.Item.update(starred=True).where(
            (models.Item.feed_id == feed_id) & (models.Item.guid == guid)
        ).execute()
    return json_response({})


@api.route('/items/unstar/multiple', methods=['PUT'])
def items_unstar_multiple():
    """Mark multiple items as unstarred."""
    items = flask.request.get_json().get('items', [])
    for i in items:
        feed_id = int(i['feedId'])
        guidHash = i['guidHash']
        guid = base64.decodebytes(guidHash.encode()).decode()
        models.Item.update(starred=False).where(
            (models.Item.feed_id == feed_id) & (models.Item.guid == guid)
        ).execute()
    return json_response({})
*/
