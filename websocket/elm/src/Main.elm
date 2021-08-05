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
    { isBannerMoving : Bool
    , currentSong : SongInfo
    , marqueeMessage : String
    , isMarqueeVisible : Bool
    }


init : () -> ( Model, Cmd Msg )
init _ =
    ( { isBannerMoving = False
      , currentSong = SongInfo "" "" ""
      , marqueeMessage = ""
      , isMarqueeVisible = False
      }
    , Process.sleep 30000 |> Task.perform (always MarqueeTick)
    )



-- UPDATE


type Msg
    = Recv String
    | StopBanner
    | MarqueeTick
    | StopMarquee


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        MarqueeTick ->
            ( { model | isMarqueeVisible = True }
            , Process.sleep 65000 |> Task.perform (always StopMarquee)
            )

        Recv message ->
            case D.decodeString websocketMessageDecoder message of
                Ok ws ->
                    case ws.action of
                        "spotify_music_updated" ->
                            case D.decodeString songInfoDecoder ws.payload of
                                Ok song ->
                                    ( { model | currentSong = song, isBannerMoving = True }
                                    , Process.sleep 15000 |> Task.perform (always StopBanner)
                                    )

                                Err _ ->
                                    ( model, Cmd.none )

                        "tts_created" ->
                            ( model, playUrl ws.payload )

                        "marquee_updated" ->
                            ( { model | marqueeMessage = ws.payload }, Cmd.none )

                        _ ->
                            ( model, Cmd.none )

                Err _ ->
                    ( model, Cmd.none )

        StopBanner ->
            ( { model | isBannerMoving = False }
            , Cmd.none
            )

        StopMarquee ->
            ( { model | isMarqueeVisible = False }
            , Process.sleep 30000 |> Task.perform (always MarqueeTick)
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
    div [ id "root" ]
        [ div
            [ classList
                [ ( "main", True )
                , ( "animate", model.isBannerMoving )
                ]
            ]
            (songInfoView model.currentSong)
        , node "marquee"
            [ attribute "scrolldelay" "60"
            , classList [ ( "animate-marquee", model.isMarqueeVisible ) ]
            ]
            [ text model.marqueeMessage ]
        ]



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
