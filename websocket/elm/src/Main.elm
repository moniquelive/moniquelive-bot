port module Main exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Json.Decode as D
import Process
import Task



-- MAIN


main : Program () Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }



-- PORTS


port playUrl : String -> Cmd msg


port messageReceiver : (String -> msg) -> Sub msg



-- MODEL


type alias WebsocketMessage =
    { action : String
    , payload : String
    }


type alias SongInfo =
    { cover : String
    , title : String
    , artist : String
    }


type alias Model =
    { isAnimating : Bool
    , currentSong : SongInfo
    }


init : () -> ( Model, Cmd Msg )
init _ =
    ( { isAnimating = False
      , currentSong = SongInfo "" "" ""
      }
    , Cmd.none
    )



-- UPDATE


type Msg
    = Recv String
    | StopAnimation


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Recv message ->
            case D.decodeString websocketMessageDecoder message of
                Ok ws ->
                    case ws.action of
                        "spotify_music_updated" ->
                            case D.decodeString songInfoDecoder ws.payload of
                                Ok song ->
                                    ( { model | currentSong = song, isAnimating = True }
                                    , Process.sleep 15000 |> Task.perform (always StopAnimation)
                                    )
                                Err _ ->
                                    ( model, Cmd.none )
                        "tts_created" ->
                            ( model, playUrl ws.payload )
                        "marquee_updated" ->
                            ( model, Cmd.none )
                        _ ->
                            ( model, Cmd.none )

                Err _ ->
                    ( model, Cmd.none )

        StopAnimation ->
            ( { model | isAnimating = False }
            , Cmd.none
            )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    messageReceiver Recv



-- VIEW


songInfoView : SongInfo -> List (Html Msg)
songInfoView song =
    [ div [ class "cover" ] [ img [ id "coverImg", src song.cover ] [] ]
    , div [ class "container" ]
        [ div [ class "title" ] [ text song.title ]
        , div [ class "artist" ] [ text song.artist ]
        ]
    ]


view : Model -> Html Msg
view model =
    div
        [ classList
            [ ( "main", True )
            , ( "animate", model.isAnimating )
            ]
        ]
        (songInfoView model.currentSong)



-- JSON decode


websocketMessageDecoder : D.Decoder WebsocketMessage
websocketMessageDecoder =
    D.map2 WebsocketMessage
        (D.field "action" D.string)
        (D.field "payload" D.string)


songInfoDecoder : D.Decoder SongInfo
songInfoDecoder =
    D.map3 SongInfo
        (D.field "imgUrl" D.string)
        (D.field "title" D.string)
        (D.field "artist" D.string)
