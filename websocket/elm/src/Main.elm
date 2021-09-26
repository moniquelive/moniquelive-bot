port module Main exposing (..)

import Animation exposing (percent, px)
import Animation.Spring.Presets exposing (stiff, wobbly)
import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Json.Decode as D
import Process
import Task
import Time



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
    { currentSong : SongInfo
    , currentSongStyle : Animation.State
    , marqueeMessage : String
    , isMarqueeVisible : Bool
    }


init : () -> ( Model, Cmd Msg )
init _ =
    ( { currentSong = SongInfo "" "" ""
      , currentSongStyle = Animation.styleWith (Animation.spring wobbly) [ Animation.translate (percent 115) (percent 0) ]
      , marqueeMessage = ""
      , isMarqueeVisible = False
      }
    , Process.sleep 30000 |> Task.perform (always MarqueeTick)
    )



-- UPDATE


type Msg
    = Recv String
    | Animate Animation.Msg
    | MarqueeTick
    | StopMarquee


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Animate animMsg ->
            let
                newCurrentSongStyle =
                    Animation.update animMsg model.currentSongStyle
            in
            ( { model | currentSongStyle = newCurrentSongStyle }, Cmd.none )

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
                                    let
                                        newCurrentSongStyle =
                                            Animation.interrupt
                                                [ Animation.to [ Animation.translate (percent 0) (percent 0) ]
                                                , Animation.wait (Time.millisToPosix <| 8 * 1000)
                                                , Animation.to [ Animation.translate (percent 115) (percent 0) ]
                                                ]
                                                model.currentSongStyle
                                    in
                                    ( { model
                                        | currentSong = song
                                        , currentSongStyle = newCurrentSongStyle
                                      }
                                    , Cmd.none
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

        StopMarquee ->
            ( { model | isMarqueeVisible = False }
            , Process.sleep 30000 |> Task.perform (always MarqueeTick)
            )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ messageReceiver Recv
        , Animation.subscription Animate [ model.currentSongStyle ]
        ]



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
            (Animation.render model.currentSongStyle
                ++ [ class "main" ]
            )
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
