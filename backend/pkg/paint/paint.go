package paint

import (
	"encoding/json"
	"image"
	"log"

	"paint.pecet.it/pkg/wsmanager"
)

type Paint struct {
	Room  *wsmanager.Room
	Image image.Image
}

// Point odpowiada interfejsowi Point z TypeScriptu
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// DrawEvent odpowiada interfejsowi DrawEvent z TypeScriptu
type DrawEvent struct {
	Action string  `json:"action"` // "start" | "draw" | "end"
	Point  Point   `json:"point"`
	Color  string  `json:"color"`
	Size   float64 `json:"size"`
	UserID string  `json:"userId,omitempty"`
}

func ImplementPaint(room *wsmanager.Room) {
	room.RegisterEventHandler("drawing", func(evt *wsmanager.Event) {
		var drawEvt DrawEvent

		// Zakładam, że evt.Payload to json.RawMessage (lub []byte).
		// Parsujemy payload do naszej struktury DrawEvent.
		err := json.Unmarshal(evt.Payload, &drawEvt)
		if err != nil {
			log.Printf("Błąd podczas parsowania eventu drawing: %v", err)
			return
		}

		// Obsługa pod-eventów na podstawie pola Action
		switch drawEvt.Action {
		case "start":
			log.Printf("Użytkownik %s ROZPOCZĄŁ rysowanie w punkcie (%f, %f) kolorem %s",
				drawEvt.UserID, drawEvt.Point.X, drawEvt.Point.Y, drawEvt.Color)
			// TODO: Tutaj możesz zainicjować stan rysowania, jeśli trzymasz go po stronie serwera

		case "draw":
			// log.Printf("Użytkownik %s rysuje w (%f, %f)", drawEvt.UserID, drawEvt.Point.X, drawEvt.Point.Y)
			// TODO: Zaktualizuj obraz (Paint.Image) o nową linię/punkt

		case "end":
			log.Printf("Użytkownik %s ZAKOŃCZYŁ rysowanie.", drawEvt.UserID)
			// TODO: Zakończ ścieżkę dla danego użytkownika

		default:
			log.Printf("Otrzymano nieznaną akcję rysowania: %s", drawEvt.Action)
			return
		}

		// Z Twojego kodu React wynika, że klient nasłuchujący odbiera bezpośrednio DrawEvent:
		// const data: DrawEvent = JSON.parse(message.data);
		// Dlatego rozsyłamy dalej (broadcast) sam payload bez zewnętrznego wrappera "type".

		broadcastData, err := json.Marshal(drawEvt)
		if err != nil {
			log.Printf("Błąd serializacji eventu do rozgłoszenia: %v", err)
			return
		}

		room.Broadcast(broadcastData)
	})
}
