package main

import (
	"fmt"

	"github.com/realPy/hogosuru"
	"github.com/realPy/hogosuru/date"
	"github.com/realPy/hogosuru/document"
	"github.com/realPy/hogosuru/documentfragment"
	"github.com/realPy/hogosuru/event"
	"github.com/realPy/hogosuru/fetch"
	"github.com/realPy/hogosuru/hogosurudebug"
	"github.com/realPy/hogosuru/htmlanchorelement"
	"github.com/realPy/hogosuru/htmlbuttonelement"
	"github.com/realPy/hogosuru/htmldivelement"
	"github.com/realPy/hogosuru/htmlelement"
	"github.com/realPy/hogosuru/htmlinputelement"
	"github.com/realPy/hogosuru/htmlspanelement"
	"github.com/realPy/hogosuru/htmltemplateelement"
	"github.com/realPy/hogosuru/keyboardevent"
	"github.com/realPy/hogosuru/messageevent"
	"github.com/realPy/hogosuru/node"
	"github.com/realPy/hogosuru/promise"
	"github.com/realPy/hogosuru/response"
	"github.com/realPy/hogosuru/websocket"
)

type GlobalContainer struct {
	parentNode     node.Node
	template       htmltemplateelement.HtmlTemplateElement
	urlWS          htmlinputelement.HtmlInputElement
	msg2Send       htmlinputelement.HtmlInputElement
	divConnect     htmldivelement.HtmlDivElement
	divMsg         htmldivelement.HtmlDivElement
	connectButton  htmlbuttonelement.HtmlButtonElement
	sendButton     htmlbuttonelement.HtmlButtonElement
	logcontent     htmldivelement.HtmlDivElement
	ws             websocket.WebSocket
	d              document.Document
	connectedIcon  htmlspanelement.HtmlSpanElement
	disconnectIcon htmlspanelement.HtmlSpanElement
	isConnected    bool
}

func ElementStatusHidden(e htmlelement.HtmlElement, status bool) {

	if status {
		e.Style_().SetProperty("display", "none")

	} else {
		e.Style_().RemoveProperty("display")
	}

}

func (w *GlobalContainer) ConnectionStatus(connected bool) {

	w.isConnected = connected
	if connected {
		ElementStatusHidden(w.connectedIcon.HtmlElement, false)
		ElementStatusHidden(w.disconnectIcon.HtmlElement, true)
		ElementStatusHidden(w.divConnect.HtmlElement, true)
		ElementStatusHidden(w.divMsg.HtmlElement, false)

	} else {
		ElementStatusHidden(w.connectedIcon.HtmlElement, true)
		ElementStatusHidden(w.disconnectIcon.HtmlElement, false)
		ElementStatusHidden(w.divConnect.HtmlElement, false)
		ElementStatusHidden(w.divMsg.HtmlElement, true)
	}

}

func (w *GlobalContainer) SetLog(typemsg string, msg string) {

	if fragment, err := w.template.Content(); hogosuru.AssertErr(err) {
		if cloneNode, err := fragment.GetElementById("logmessage"); hogosuru.AssertErr(err) {
			if clone, err := w.d.ImportNode(cloneNode.Node, true); hogosuru.AssertErr(err) {

				if div, ok := clone.(htmldivelement.HtmlDivElement); ok {

					if typemsg == "error" {
						div.SetClassName("logmessage logmessageerror")
					}

					if divtext, err := div.QuerySelector("#logmessagetxt"); hogosuru.AssertErr(err) {

						divtext.SetTextContent(msg)

					}

					if divtime, err := div.QuerySelector("#logmessagetime"); hogosuru.AssertErr(err) {
						if d, err := date.New(); hogosuru.AssertErr(err) {

							dh, _ := d.GetHours()
							dm, _ := d.GetMinutes()
							ds, _ := d.GetSeconds()
							divtime.SetTextContent(fmt.Sprintf("%02d:%02d:%02d", dh, dm, ds))

						}

					}

					w.logcontent.AppendChild(div.Node)
					if scrollHeight, _ := w.logcontent.ScrollHeight(); hogosuru.AssertErr(err) {
						w.logcontent.SetScrollTop(scrollHeight)
					}

				}
			}
		}
	}
}

func (w *GlobalContainer) InstallWS(url string) {
	var err error
	if w.ws, err = websocket.New(url); err == nil {
		w.ws.SetOnOpen(func(e event.Event) {

			w.SetLog("info", "Successful connected!")
			w.divConnect.Style_().SetProperty("display", "none")
			w.divMsg.Style_().RemoveProperty("display")
			w.ConnectionStatus(true)

		})

		w.ws.SetOnMessage(func(e messageevent.MessageEvent) {

			if message, err := e.Data(); hogosuru.AssertErr(err) {
				if s, ok := message.(string); ok {
					w.SetLog("info", "Server << "+s)
				} else {
					w.SetLog("error", "Receive a no text data from server")
				}

			}

		})

		w.ws.SetOnClose(func(e event.Event) {
			if w.isConnected == false {
				w.SetLog("error", "Unable to connected to server")
			} else {
				w.SetLog("error", "Webssocket Closed!")
			}
			w.ConnectionStatus(false)

			w.divMsg.Style_().SetProperty("display", "none")

			w.divConnect.Style_().RemoveProperty("display")
			w.ws = websocket.WebSocket{}

		})
	} else {
		w.SetLog("error", err.Error())
	}

}

func (w *GlobalContainer) OnLoad(d document.Document, n node.Node, route string) (*promise.Promise, []hogosuru.Rendering) {

	htmlinputelement.GetInterface()
	htmlbuttonelement.GetInterface()
	htmltemplateelement.GetInterface()
	documentfragment.GetInterface()
	htmldivelement.GetInterface()
	htmlspanelement.GetInterface()
	htmlanchorelement.GetInterface()
	keyboardevent.GetInterface()

	w.parentNode = n
	w.d = d
	var ret *promise.Promise

	if f, err := fetch.New("main.html"); hogosuru.AssertErr(err) {
		textpromise, _ := f.Then(func(r response.Response) *promise.Promise {

			if promise, err := r.Text(); hogosuru.AssertErr(err) {
				return &promise
			}

			return nil

		}, nil)

		textpromise.Then(func(i interface{}) *promise.Promise {

			if element, err := d.DocumentElement(); hogosuru.AssertErr(err) {
				element.SetInnerHTML(i.(string))

				if addrwsText, err := d.GetElementById("wsserver"); hogosuru.AssertErr(err) {

					if elem, err := addrwsText.Discover(); hogosuru.AssertErr(err) {

						if input, ok := elem.(htmlinputelement.HtmlInputElement); ok {
							w.urlWS = input

						}
					}

				}

				if msgText, err := d.GetElementById("msg2send"); hogosuru.AssertErr(err) {

					if elem, err := msgText.Discover(); hogosuru.AssertErr(err) {

						if input, ok := elem.(htmlinputelement.HtmlInputElement); ok {
							w.msg2Send = input

						}
					}

				}

				if elemButton, err := d.GetElementById("buttonconnect"); hogosuru.AssertErr(err) {

					if elem, err := elemButton.Discover(); hogosuru.AssertErr(err) {

						if button, ok := elem.(htmlbuttonelement.HtmlButtonElement); ok {
							w.connectButton = button
							w.connectButton.OnClick(func(e event.Event) {

								if w.ws.Empty() {

									if url, err := w.urlWS.Value(); hogosuru.AssertErr(err) {
										if len(url) == 0 {
											w.InstallWS("wss://ws.ifelse.io")
										} else {
											w.InstallWS(url)
										}

									}

								}

							})

						}
					}

				}
				if sendButton, err := d.GetElementById("sendmsg"); hogosuru.AssertErr(err) {

					if elem, err := sendButton.Discover(); hogosuru.AssertErr(err) {

						if button, ok := elem.(htmlbuttonelement.HtmlButtonElement); ok {
							w.sendButton = button
							w.sendButton.OnClick(func(e event.Event) {

								if !w.ws.Empty() {

									if msg, err := w.msg2Send.Value(); hogosuru.AssertErr(err) {
										if len(msg) > 0 {
											w.SetLog("info", "YOU>>"+msg)
											if errsend := w.ws.Send(msg); errsend != nil {
												w.SetLog("error", errsend.Error())
											}
											hogosuru.AssertErr(w.msg2Send.SetValue(""))
										}

									}

								}

							})

						}
					}

				}

				w.msg2Send.AddEventListener("keyup", func(e event.Event) {

					if kei, err := e.Discover(); hogosuru.AssertErr(err) {
						if ke, ok := kei.(keyboardevent.KeyboardEventFrom); ok {

							if ckey, err := ke.KeyboardEvent_().Key(); hogosuru.AssertErr(err) {
								if ckey == "Enter" {
									hogosuru.AssertErr(w.sendButton.Click())

								}

							}
						}
					}

				})

				if elemTemplateMsg, err := d.GetElementById("logmessagetemplate"); hogosuru.AssertErr(err) {

					if elemTemplateInstance, err := elemTemplateMsg.Discover(); hogosuru.AssertErr(err) {

						if t, ok := elemTemplateInstance.(htmltemplateelement.HtmlTemplateElement); ok {
							w.template = t
						}
					}

				}

				if elemLogContent, err := d.GetElementById("logcontent"); hogosuru.AssertErr(err) {

					if elemLogContentInstance, err := elemLogContent.Discover(); hogosuru.AssertErr(err) {

						if l, ok := elemLogContentInstance.(htmldivelement.HtmlDivElement); ok {
							w.logcontent = l
						}
					}

				}

				if elemConnect, err := d.GetElementById("connect"); hogosuru.AssertErr(err) {

					if elemConnectInstance, err := elemConnect.Discover(); hogosuru.AssertErr(err) {

						if e, ok := elemConnectInstance.(htmldivelement.HtmlDivElement); ok {
							w.divConnect = e
						}
					}

				}

				if elemMsg, err := d.GetElementById("msg"); hogosuru.AssertErr(err) {

					if elemMsgInstance, err := elemMsg.Discover(); hogosuru.AssertErr(err) {

						if e, ok := elemMsgInstance.(htmldivelement.HtmlDivElement); ok {
							w.divMsg = e
							e.Style_().SetProperty("display", "none")
						}
					}

				}

				if elemcicon, err := d.GetElementById("iconconnect"); hogosuru.AssertErr(err) {

					if elemciconInstance, err := elemcicon.Discover(); hogosuru.AssertErr(err) {

						if e, ok := elemciconInstance.(htmlspanelement.HtmlSpanElement); ok {
							w.connectedIcon = e

						}
					}

				}

				if elemdicon, err := d.GetElementById("icondisconnect"); hogosuru.AssertErr(err) {

					if elemdiconInstance, err := elemdicon.Discover(); hogosuru.AssertErr(err) {

						if e, ok := elemdiconInstance.(htmlspanelement.HtmlSpanElement); ok {
							w.disconnectIcon = e

						}
					}

				}

				if elemlink, err := d.GetElementById("ws-disconnect"); hogosuru.AssertErr(err) {

					if elemlinkInstance, err := elemlink.Discover(); hogosuru.AssertErr(err) {

						if e, ok := elemlinkInstance.(htmlanchorelement.HtmlAnchorElement); ok {

							e.OnClick(func(e event.Event) {
								if !w.ws.Empty() {
									w.ws.Close()
								}
								e.PreventDefault()
							})

						}
					}

				}

				if elemlink, err := d.GetElementById("trashlog"); hogosuru.AssertErr(err) {

					if elemlinkInstance, err := elemlink.Discover(); hogosuru.AssertErr(err) {

						if e, ok := elemlinkInstance.(htmlanchorelement.HtmlAnchorElement); ok {

							e.OnClick(func(e event.Event) {
								for r, err := w.logcontent.FirstChild(); err == nil; r, err = w.logcontent.FirstChild() {
									w.logcontent.RemoveChild(r)
								}
								e.PreventDefault()
							})

						}
					}

				}

				w.ConnectionStatus(false)

			}

			return nil
		}, nil)

		ret = &textpromise

	}

	return ret, nil
}

func (w *GlobalContainer) Node(r hogosuru.Rendering) node.Node {

	return w.parentNode
}

func (w *GlobalContainer) OnEndChildRendering(r hogosuru.Rendering) {

}

func (w *GlobalContainer) OnEndChildsRendering() {

}

func (w *GlobalContainer) OnUnload() {

}

func main() {
	hogosuru.Init()
	hogosurudebug.EnableDebug()
	hogosuru.Router().DefaultRendering(&GlobalContainer{})
	hogosuru.Router().Start(hogosuru.HASHROUTE)
	ch := make(chan struct{})
	<-ch

}
