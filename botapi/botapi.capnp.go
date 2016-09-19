package botapi

// AUTO GENERATED - DO NOT EDIT

import (
	context "golang.org/x/net/context"
	strconv "strconv"
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
	server "zombiezen.com/go/capnproto2/server"
)

type AiConnector struct{ Client capnp.Client }

func (c AiConnector) Connect(ctx context.Context, params func(ConnectRequest) error, opts ...capnp.CallOption) AiConnector_connect_Results_Promise {
	if c.Client == nil {
		return AiConnector_connect_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0x9804b41cc3cba212,
			MethodID:      0,
			InterfaceName: "botapi.capnp:AiConnector",
			MethodName:    "connect",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 2}
		call.ParamsFunc = func(s capnp.Struct) error { return params(ConnectRequest{Struct: s}) }
	}
	return AiConnector_connect_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type AiConnector_Server interface {
	Connect(AiConnector_connect) error
}

func AiConnector_ServerToClient(s AiConnector_Server) AiConnector {
	c, _ := s.(server.Closer)
	return AiConnector{Client: server.New(AiConnector_Methods(nil, s), c)}
}

func AiConnector_Methods(methods []server.Method, s AiConnector_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0x9804b41cc3cba212,
			MethodID:      0,
			InterfaceName: "botapi.capnp:AiConnector",
			MethodName:    "connect",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := AiConnector_connect{c, opts, ConnectRequest{Struct: p}, AiConnector_connect_Results{Struct: r}}
			return s.Connect(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 0},
	})

	return methods
}

// AiConnector_connect holds the arguments for a server call to AiConnector.connect.
type AiConnector_connect struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  ConnectRequest
	Results AiConnector_connect_Results
}

type AiConnector_connect_Results struct{ capnp.Struct }

// AiConnector_connect_Results_TypeID is the unique identifier for the type AiConnector_connect_Results.
const AiConnector_connect_Results_TypeID = 0xaf821edee86a29e4

func NewAiConnector_connect_Results(s *capnp.Segment) (AiConnector_connect_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return AiConnector_connect_Results{st}, err
}

func NewRootAiConnector_connect_Results(s *capnp.Segment) (AiConnector_connect_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0})
	return AiConnector_connect_Results{st}, err
}

func ReadRootAiConnector_connect_Results(msg *capnp.Message) (AiConnector_connect_Results, error) {
	root, err := msg.RootPtr()
	return AiConnector_connect_Results{root.Struct()}, err
}

func (s AiConnector_connect_Results) String() string {
	str, _ := text.Marshal(0xaf821edee86a29e4, s.Struct)
	return str
}

// AiConnector_connect_Results_List is a list of AiConnector_connect_Results.
type AiConnector_connect_Results_List struct{ capnp.List }

// NewAiConnector_connect_Results creates a new list of AiConnector_connect_Results.
func NewAiConnector_connect_Results_List(s *capnp.Segment, sz int32) (AiConnector_connect_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 0}, sz)
	return AiConnector_connect_Results_List{l}, err
}

func (s AiConnector_connect_Results_List) At(i int) AiConnector_connect_Results {
	return AiConnector_connect_Results{s.List.Struct(i)}
}

func (s AiConnector_connect_Results_List) Set(i int, v AiConnector_connect_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// AiConnector_connect_Results_Promise is a wrapper for a AiConnector_connect_Results promised by a client call.
type AiConnector_connect_Results_Promise struct{ *capnp.Pipeline }

func (p AiConnector_connect_Results_Promise) Struct() (AiConnector_connect_Results, error) {
	s, err := p.Pipeline.Struct()
	return AiConnector_connect_Results{s}, err
}

type ConnectRequest struct{ capnp.Struct }

// ConnectRequest_TypeID is the unique identifier for the type ConnectRequest.
const ConnectRequest_TypeID = 0x95f2e57bf5bcea49

func NewConnectRequest(s *capnp.Segment) (ConnectRequest, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return ConnectRequest{st}, err
}

func NewRootConnectRequest(s *capnp.Segment) (ConnectRequest, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return ConnectRequest{st}, err
}

func ReadRootConnectRequest(msg *capnp.Message) (ConnectRequest, error) {
	root, err := msg.RootPtr()
	return ConnectRequest{root.Struct()}, err
}

func (s ConnectRequest) String() string {
	str, _ := text.Marshal(0x95f2e57bf5bcea49, s.Struct)
	return str
}

func (s ConnectRequest) Credentials() (Credentials, error) {
	p, err := s.Struct.Ptr(0)
	return Credentials{Struct: p.Struct()}, err
}

func (s ConnectRequest) HasCredentials() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s ConnectRequest) SetCredentials(v Credentials) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewCredentials sets the credentials field to a newly
// allocated Credentials struct, preferring placement in s's segment.
func (s ConnectRequest) NewCredentials() (Credentials, error) {
	ss, err := NewCredentials(s.Struct.Segment())
	if err != nil {
		return Credentials{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s ConnectRequest) Ai() Ai {
	p, _ := s.Struct.Ptr(1)
	return Ai{Client: p.Interface().Client()}
}

func (s ConnectRequest) HasAi() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s ConnectRequest) SetAi(v Ai) error {
	if v.Client == nil {
		return s.Struct.SetPtr(1, capnp.Ptr{})
	}
	seg := s.Segment()
	in := capnp.NewInterface(seg, seg.Message().AddCap(v.Client))
	return s.Struct.SetPtr(1, in.ToPtr())
}

// ConnectRequest_List is a list of ConnectRequest.
type ConnectRequest_List struct{ capnp.List }

// NewConnectRequest creates a new list of ConnectRequest.
func NewConnectRequest_List(s *capnp.Segment, sz int32) (ConnectRequest_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return ConnectRequest_List{l}, err
}

func (s ConnectRequest_List) At(i int) ConnectRequest { return ConnectRequest{s.List.Struct(i)} }

func (s ConnectRequest_List) Set(i int, v ConnectRequest) error { return s.List.SetStruct(i, v.Struct) }

// ConnectRequest_Promise is a wrapper for a ConnectRequest promised by a client call.
type ConnectRequest_Promise struct{ *capnp.Pipeline }

func (p ConnectRequest_Promise) Struct() (ConnectRequest, error) {
	s, err := p.Pipeline.Struct()
	return ConnectRequest{s}, err
}

func (p ConnectRequest_Promise) Credentials() Credentials_Promise {
	return Credentials_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p ConnectRequest_Promise) Ai() Ai {
	return Ai{Client: p.Pipeline.GetPipeline(1).Client()}
}

type Credentials struct{ capnp.Struct }

// Credentials_TypeID is the unique identifier for the type Credentials.
const Credentials_TypeID = 0xcca8fe75a57f1ea7

func NewCredentials(s *capnp.Segment) (Credentials, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Credentials{st}, err
}

func NewRootCredentials(s *capnp.Segment) (Credentials, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Credentials{st}, err
}

func ReadRootCredentials(msg *capnp.Message) (Credentials, error) {
	root, err := msg.RootPtr()
	return Credentials{root.Struct()}, err
}

func (s Credentials) String() string {
	str, _ := text.Marshal(0xcca8fe75a57f1ea7, s.Struct)
	return str
}

func (s Credentials) SecretToken() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s Credentials) HasSecretToken() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Credentials) SecretTokenBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s Credentials) SetSecretToken(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s Credentials) BotName() (string, error) {
	p, err := s.Struct.Ptr(1)
	return p.Text(), err
}

func (s Credentials) HasBotName() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Credentials) BotNameBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(1)
	return p.TextBytes(), err
}

func (s Credentials) SetBotName(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(1, t.List.ToPtr())
}

// Credentials_List is a list of Credentials.
type Credentials_List struct{ capnp.List }

// NewCredentials creates a new list of Credentials.
func NewCredentials_List(s *capnp.Segment, sz int32) (Credentials_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return Credentials_List{l}, err
}

func (s Credentials_List) At(i int) Credentials { return Credentials{s.List.Struct(i)} }

func (s Credentials_List) Set(i int, v Credentials) error { return s.List.SetStruct(i, v.Struct) }

// Credentials_Promise is a wrapper for a Credentials promised by a client call.
type Credentials_Promise struct{ *capnp.Pipeline }

func (p Credentials_Promise) Struct() (Credentials, error) {
	s, err := p.Pipeline.Struct()
	return Credentials{s}, err
}

type Ai struct{ Client capnp.Client }

func (c Ai) TakeTurn(ctx context.Context, params func(Ai_takeTurn_Params) error, opts ...capnp.CallOption) Ai_takeTurn_Results_Promise {
	if c.Client == nil {
		return Ai_takeTurn_Results_Promise{Pipeline: capnp.NewPipeline(capnp.ErrorAnswer(capnp.ErrNullClient))}
	}
	call := &capnp.Call{
		Ctx: ctx,
		Method: capnp.Method{
			InterfaceID:   0xd403ce7bb5b69f1f,
			MethodID:      0,
			InterfaceName: "botapi.capnp:Ai",
			MethodName:    "takeTurn",
		},
		Options: capnp.NewCallOptions(opts),
	}
	if params != nil {
		call.ParamsSize = capnp.ObjectSize{DataSize: 0, PointerCount: 1}
		call.ParamsFunc = func(s capnp.Struct) error { return params(Ai_takeTurn_Params{Struct: s}) }
	}
	return Ai_takeTurn_Results_Promise{Pipeline: capnp.NewPipeline(c.Client.Call(call))}
}

type Ai_Server interface {
	TakeTurn(Ai_takeTurn) error
}

func Ai_ServerToClient(s Ai_Server) Ai {
	c, _ := s.(server.Closer)
	return Ai{Client: server.New(Ai_Methods(nil, s), c)}
}

func Ai_Methods(methods []server.Method, s Ai_Server) []server.Method {
	if cap(methods) == 0 {
		methods = make([]server.Method, 0, 1)
	}

	methods = append(methods, server.Method{
		Method: capnp.Method{
			InterfaceID:   0xd403ce7bb5b69f1f,
			MethodID:      0,
			InterfaceName: "botapi.capnp:Ai",
			MethodName:    "takeTurn",
		},
		Impl: func(c context.Context, opts capnp.CallOptions, p, r capnp.Struct) error {
			call := Ai_takeTurn{c, opts, Ai_takeTurn_Params{Struct: p}, Ai_takeTurn_Results{Struct: r}}
			return s.TakeTurn(call)
		},
		ResultsSize: capnp.ObjectSize{DataSize: 0, PointerCount: 1},
	})

	return methods
}

// Ai_takeTurn holds the arguments for a server call to Ai.takeTurn.
type Ai_takeTurn struct {
	Ctx     context.Context
	Options capnp.CallOptions
	Params  Ai_takeTurn_Params
	Results Ai_takeTurn_Results
}

type Ai_takeTurn_Params struct{ capnp.Struct }

// Ai_takeTurn_Params_TypeID is the unique identifier for the type Ai_takeTurn_Params.
const Ai_takeTurn_Params_TypeID = 0x91b9eb0bc884d7fb

func NewAi_takeTurn_Params(s *capnp.Segment) (Ai_takeTurn_Params, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Ai_takeTurn_Params{st}, err
}

func NewRootAi_takeTurn_Params(s *capnp.Segment) (Ai_takeTurn_Params, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Ai_takeTurn_Params{st}, err
}

func ReadRootAi_takeTurn_Params(msg *capnp.Message) (Ai_takeTurn_Params, error) {
	root, err := msg.RootPtr()
	return Ai_takeTurn_Params{root.Struct()}, err
}

func (s Ai_takeTurn_Params) String() string {
	str, _ := text.Marshal(0x91b9eb0bc884d7fb, s.Struct)
	return str
}

func (s Ai_takeTurn_Params) Board() (InitialBoard, error) {
	p, err := s.Struct.Ptr(0)
	return InitialBoard{Struct: p.Struct()}, err
}

func (s Ai_takeTurn_Params) HasBoard() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Ai_takeTurn_Params) SetBoard(v InitialBoard) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewBoard sets the board field to a newly
// allocated InitialBoard struct, preferring placement in s's segment.
func (s Ai_takeTurn_Params) NewBoard() (InitialBoard, error) {
	ss, err := NewInitialBoard(s.Struct.Segment())
	if err != nil {
		return InitialBoard{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// Ai_takeTurn_Params_List is a list of Ai_takeTurn_Params.
type Ai_takeTurn_Params_List struct{ capnp.List }

// NewAi_takeTurn_Params creates a new list of Ai_takeTurn_Params.
func NewAi_takeTurn_Params_List(s *capnp.Segment, sz int32) (Ai_takeTurn_Params_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Ai_takeTurn_Params_List{l}, err
}

func (s Ai_takeTurn_Params_List) At(i int) Ai_takeTurn_Params {
	return Ai_takeTurn_Params{s.List.Struct(i)}
}

func (s Ai_takeTurn_Params_List) Set(i int, v Ai_takeTurn_Params) error {
	return s.List.SetStruct(i, v.Struct)
}

// Ai_takeTurn_Params_Promise is a wrapper for a Ai_takeTurn_Params promised by a client call.
type Ai_takeTurn_Params_Promise struct{ *capnp.Pipeline }

func (p Ai_takeTurn_Params_Promise) Struct() (Ai_takeTurn_Params, error) {
	s, err := p.Pipeline.Struct()
	return Ai_takeTurn_Params{s}, err
}

func (p Ai_takeTurn_Params_Promise) Board() InitialBoard_Promise {
	return InitialBoard_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type Ai_takeTurn_Results struct{ capnp.Struct }

// Ai_takeTurn_Results_TypeID is the unique identifier for the type Ai_takeTurn_Results.
const Ai_takeTurn_Results_TypeID = 0x8d265c88e8a2e488

func NewAi_takeTurn_Results(s *capnp.Segment) (Ai_takeTurn_Results, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Ai_takeTurn_Results{st}, err
}

func NewRootAi_takeTurn_Results(s *capnp.Segment) (Ai_takeTurn_Results, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return Ai_takeTurn_Results{st}, err
}

func ReadRootAi_takeTurn_Results(msg *capnp.Message) (Ai_takeTurn_Results, error) {
	root, err := msg.RootPtr()
	return Ai_takeTurn_Results{root.Struct()}, err
}

func (s Ai_takeTurn_Results) String() string {
	str, _ := text.Marshal(0x8d265c88e8a2e488, s.Struct)
	return str
}

func (s Ai_takeTurn_Results) Turns() (Turn_List, error) {
	p, err := s.Struct.Ptr(0)
	return Turn_List{List: p.List()}, err
}

func (s Ai_takeTurn_Results) HasTurns() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Ai_takeTurn_Results) SetTurns(v Turn_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewTurns sets the turns field to a newly
// allocated Turn_List, preferring placement in s's segment.
func (s Ai_takeTurn_Results) NewTurns(n int32) (Turn_List, error) {
	l, err := NewTurn_List(s.Struct.Segment(), n)
	if err != nil {
		return Turn_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

// Ai_takeTurn_Results_List is a list of Ai_takeTurn_Results.
type Ai_takeTurn_Results_List struct{ capnp.List }

// NewAi_takeTurn_Results creates a new list of Ai_takeTurn_Results.
func NewAi_takeTurn_Results_List(s *capnp.Segment, sz int32) (Ai_takeTurn_Results_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return Ai_takeTurn_Results_List{l}, err
}

func (s Ai_takeTurn_Results_List) At(i int) Ai_takeTurn_Results {
	return Ai_takeTurn_Results{s.List.Struct(i)}
}

func (s Ai_takeTurn_Results_List) Set(i int, v Ai_takeTurn_Results) error {
	return s.List.SetStruct(i, v.Struct)
}

// Ai_takeTurn_Results_Promise is a wrapper for a Ai_takeTurn_Results promised by a client call.
type Ai_takeTurn_Results_Promise struct{ *capnp.Pipeline }

func (p Ai_takeTurn_Results_Promise) Struct() (Ai_takeTurn_Results, error) {
	s, err := p.Pipeline.Struct()
	return Ai_takeTurn_Results{s}, err
}

type Board struct{ capnp.Struct }

// Board_TypeID is the unique identifier for the type Board.
const Board_TypeID = 0xd57da3828ebb699b

func NewBoard(s *capnp.Segment) (Board, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2})
	return Board{st}, err
}

func NewRootBoard(s *capnp.Segment) (Board, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2})
	return Board{st}, err
}

func ReadRootBoard(msg *capnp.Message) (Board, error) {
	root, err := msg.RootPtr()
	return Board{root.Struct()}, err
}

func (s Board) String() string {
	str, _ := text.Marshal(0xd57da3828ebb699b, s.Struct)
	return str
}

func (s Board) GameId() (string, error) {
	p, err := s.Struct.Ptr(1)
	return p.Text(), err
}

func (s Board) HasGameId() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Board) GameIdBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(1)
	return p.TextBytes(), err
}

func (s Board) SetGameId(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(1, t.List.ToPtr())
}

func (s Board) Width() uint16 {
	return s.Struct.Uint16(0)
}

func (s Board) SetWidth(v uint16) {
	s.Struct.SetUint16(0, v)
}

func (s Board) Height() uint16 {
	return s.Struct.Uint16(2)
}

func (s Board) SetHeight(v uint16) {
	s.Struct.SetUint16(2, v)
}

func (s Board) Robots() (Robot_List, error) {
	p, err := s.Struct.Ptr(0)
	return Robot_List{List: p.List()}, err
}

func (s Board) HasRobots() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Board) SetRobots(v Robot_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewRobots sets the robots field to a newly
// allocated Robot_List, preferring placement in s's segment.
func (s Board) NewRobots(n int32) (Robot_List, error) {
	l, err := NewRobot_List(s.Struct.Segment(), n)
	if err != nil {
		return Robot_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Board) Round() int32 {
	return int32(s.Struct.Uint32(4))
}

func (s Board) SetRound(v int32) {
	s.Struct.SetUint32(4, uint32(v))
}

// Board_List is a list of Board.
type Board_List struct{ capnp.List }

// NewBoard creates a new list of Board.
func NewBoard_List(s *capnp.Segment, sz int32) (Board_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 2}, sz)
	return Board_List{l}, err
}

func (s Board_List) At(i int) Board { return Board{s.List.Struct(i)} }

func (s Board_List) Set(i int, v Board) error { return s.List.SetStruct(i, v.Struct) }

// Board_Promise is a wrapper for a Board promised by a client call.
type Board_Promise struct{ *capnp.Pipeline }

func (p Board_Promise) Struct() (Board, error) {
	s, err := p.Pipeline.Struct()
	return Board{s}, err
}

type InitialBoard struct{ capnp.Struct }

// InitialBoard_TypeID is the unique identifier for the type InitialBoard.
const InitialBoard_TypeID = 0xa01831bb8bf68e89

func NewInitialBoard(s *capnp.Segment) (InitialBoard, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return InitialBoard{st}, err
}

func NewRootInitialBoard(s *capnp.Segment) (InitialBoard, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return InitialBoard{st}, err
}

func ReadRootInitialBoard(msg *capnp.Message) (InitialBoard, error) {
	root, err := msg.RootPtr()
	return InitialBoard{root.Struct()}, err
}

func (s InitialBoard) String() string {
	str, _ := text.Marshal(0xa01831bb8bf68e89, s.Struct)
	return str
}

func (s InitialBoard) Board() (Board, error) {
	p, err := s.Struct.Ptr(0)
	return Board{Struct: p.Struct()}, err
}

func (s InitialBoard) HasBoard() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s InitialBoard) SetBoard(v Board) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewBoard sets the board field to a newly
// allocated Board struct, preferring placement in s's segment.
func (s InitialBoard) NewBoard() (Board, error) {
	ss, err := NewBoard(s.Struct.Segment())
	if err != nil {
		return Board{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s InitialBoard) Cells() (CellType_List, error) {
	p, err := s.Struct.Ptr(1)
	return CellType_List{List: p.List()}, err
}

func (s InitialBoard) HasCells() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s InitialBoard) SetCells(v CellType_List) error {
	return s.Struct.SetPtr(1, v.List.ToPtr())
}

// NewCells sets the cells field to a newly
// allocated CellType_List, preferring placement in s's segment.
func (s InitialBoard) NewCells(n int32) (CellType_List, error) {
	l, err := NewCellType_List(s.Struct.Segment(), n)
	if err != nil {
		return CellType_List{}, err
	}
	err = s.Struct.SetPtr(1, l.List.ToPtr())
	return l, err
}

// InitialBoard_List is a list of InitialBoard.
type InitialBoard_List struct{ capnp.List }

// NewInitialBoard creates a new list of InitialBoard.
func NewInitialBoard_List(s *capnp.Segment, sz int32) (InitialBoard_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return InitialBoard_List{l}, err
}

func (s InitialBoard_List) At(i int) InitialBoard { return InitialBoard{s.List.Struct(i)} }

func (s InitialBoard_List) Set(i int, v InitialBoard) error { return s.List.SetStruct(i, v.Struct) }

// InitialBoard_Promise is a wrapper for a InitialBoard promised by a client call.
type InitialBoard_Promise struct{ *capnp.Pipeline }

func (p InitialBoard_Promise) Struct() (InitialBoard, error) {
	s, err := p.Pipeline.Struct()
	return InitialBoard{s}, err
}

func (p InitialBoard_Promise) Board() Board_Promise {
	return Board_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type Robot struct{ capnp.Struct }

// Robot_TypeID is the unique identifier for the type Robot.
const Robot_TypeID = 0xa1f5501bdc903810

func NewRobot(s *capnp.Segment) (Robot, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 0})
	return Robot{st}, err
}

func NewRootRobot(s *capnp.Segment) (Robot, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 0})
	return Robot{st}, err
}

func ReadRootRobot(msg *capnp.Message) (Robot, error) {
	root, err := msg.RootPtr()
	return Robot{root.Struct()}, err
}

func (s Robot) String() string {
	str, _ := text.Marshal(0xa1f5501bdc903810, s.Struct)
	return str
}

func (s Robot) Id() uint32 {
	return s.Struct.Uint32(0)
}

func (s Robot) SetId(v uint32) {
	s.Struct.SetUint32(0, v)
}

func (s Robot) X() uint16 {
	return s.Struct.Uint16(4)
}

func (s Robot) SetX(v uint16) {
	s.Struct.SetUint16(4, v)
}

func (s Robot) Y() uint16 {
	return s.Struct.Uint16(6)
}

func (s Robot) SetY(v uint16) {
	s.Struct.SetUint16(6, v)
}

func (s Robot) Health() int16 {
	return int16(s.Struct.Uint16(8))
}

func (s Robot) SetHealth(v int16) {
	s.Struct.SetUint16(8, uint16(v))
}

func (s Robot) Faction() Faction {
	return Faction(s.Struct.Uint16(10))
}

func (s Robot) SetFaction(v Faction) {
	s.Struct.SetUint16(10, uint16(v))
}

// Robot_List is a list of Robot.
type Robot_List struct{ capnp.List }

// NewRobot creates a new list of Robot.
func NewRobot_List(s *capnp.Segment, sz int32) (Robot_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 16, PointerCount: 0}, sz)
	return Robot_List{l}, err
}

func (s Robot_List) At(i int) Robot { return Robot{s.List.Struct(i)} }

func (s Robot_List) Set(i int, v Robot) error { return s.List.SetStruct(i, v.Struct) }

// Robot_Promise is a wrapper for a Robot promised by a client call.
type Robot_Promise struct{ *capnp.Pipeline }

func (p Robot_Promise) Struct() (Robot, error) {
	s, err := p.Pipeline.Struct()
	return Robot{s}, err
}

type Replay struct{ capnp.Struct }

// Replay_TypeID is the unique identifier for the type Replay.
const Replay_TypeID = 0xb1b85070ccf68de1

func NewReplay(s *capnp.Segment) (Replay, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return Replay{st}, err
}

func NewRootReplay(s *capnp.Segment) (Replay, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return Replay{st}, err
}

func ReadRootReplay(msg *capnp.Message) (Replay, error) {
	root, err := msg.RootPtr()
	return Replay{root.Struct()}, err
}

func (s Replay) String() string {
	str, _ := text.Marshal(0xb1b85070ccf68de1, s.Struct)
	return str
}

func (s Replay) GameId() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s Replay) HasGameId() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Replay) GameIdBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s Replay) SetGameId(v string) error {
	t, err := capnp.NewText(s.Struct.Segment(), v)
	if err != nil {
		return err
	}
	return s.Struct.SetPtr(0, t.List.ToPtr())
}

func (s Replay) Initial() (InitialBoard, error) {
	p, err := s.Struct.Ptr(1)
	return InitialBoard{Struct: p.Struct()}, err
}

func (s Replay) HasInitial() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Replay) SetInitial(v InitialBoard) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewInitial sets the initial field to a newly
// allocated InitialBoard struct, preferring placement in s's segment.
func (s Replay) NewInitial() (InitialBoard, error) {
	ss, err := NewInitialBoard(s.Struct.Segment())
	if err != nil {
		return InitialBoard{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

func (s Replay) Rounds() (Replay_Round_List, error) {
	p, err := s.Struct.Ptr(2)
	return Replay_Round_List{List: p.List()}, err
}

func (s Replay) HasRounds() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s Replay) SetRounds(v Replay_Round_List) error {
	return s.Struct.SetPtr(2, v.List.ToPtr())
}

// NewRounds sets the rounds field to a newly
// allocated Replay_Round_List, preferring placement in s's segment.
func (s Replay) NewRounds(n int32) (Replay_Round_List, error) {
	l, err := NewReplay_Round_List(s.Struct.Segment(), n)
	if err != nil {
		return Replay_Round_List{}, err
	}
	err = s.Struct.SetPtr(2, l.List.ToPtr())
	return l, err
}

// Replay_List is a list of Replay.
type Replay_List struct{ capnp.List }

// NewReplay creates a new list of Replay.
func NewReplay_List(s *capnp.Segment, sz int32) (Replay_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3}, sz)
	return Replay_List{l}, err
}

func (s Replay_List) At(i int) Replay { return Replay{s.List.Struct(i)} }

func (s Replay_List) Set(i int, v Replay) error { return s.List.SetStruct(i, v.Struct) }

// Replay_Promise is a wrapper for a Replay promised by a client call.
type Replay_Promise struct{ *capnp.Pipeline }

func (p Replay_Promise) Struct() (Replay, error) {
	s, err := p.Pipeline.Struct()
	return Replay{s}, err
}

func (p Replay_Promise) Initial() InitialBoard_Promise {
	return InitialBoard_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type Replay_Round struct{ capnp.Struct }

// Replay_Round_TypeID is the unique identifier for the type Replay_Round.
const Replay_Round_TypeID = 0xa37a83b5e914a8c4

func NewReplay_Round(s *capnp.Segment) (Replay_Round, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Replay_Round{st}, err
}

func NewRootReplay_Round(s *capnp.Segment) (Replay_Round, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return Replay_Round{st}, err
}

func ReadRootReplay_Round(msg *capnp.Message) (Replay_Round, error) {
	root, err := msg.RootPtr()
	return Replay_Round{root.Struct()}, err
}

func (s Replay_Round) String() string {
	str, _ := text.Marshal(0xa37a83b5e914a8c4, s.Struct)
	return str
}

func (s Replay_Round) Moves() (Turn_List, error) {
	p, err := s.Struct.Ptr(0)
	return Turn_List{List: p.List()}, err
}

func (s Replay_Round) HasMoves() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Replay_Round) SetMoves(v Turn_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewMoves sets the moves field to a newly
// allocated Turn_List, preferring placement in s's segment.
func (s Replay_Round) NewMoves(n int32) (Turn_List, error) {
	l, err := NewTurn_List(s.Struct.Segment(), n)
	if err != nil {
		return Turn_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Replay_Round) EndBoard() (Board, error) {
	p, err := s.Struct.Ptr(1)
	return Board{Struct: p.Struct()}, err
}

func (s Replay_Round) HasEndBoard() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Replay_Round) SetEndBoard(v Board) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewEndBoard sets the endBoard field to a newly
// allocated Board struct, preferring placement in s's segment.
func (s Replay_Round) NewEndBoard() (Board, error) {
	ss, err := NewBoard(s.Struct.Segment())
	if err != nil {
		return Board{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

// Replay_Round_List is a list of Replay_Round.
type Replay_Round_List struct{ capnp.List }

// NewReplay_Round creates a new list of Replay_Round.
func NewReplay_Round_List(s *capnp.Segment, sz int32) (Replay_Round_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return Replay_Round_List{l}, err
}

func (s Replay_Round_List) At(i int) Replay_Round { return Replay_Round{s.List.Struct(i)} }

func (s Replay_Round_List) Set(i int, v Replay_Round) error { return s.List.SetStruct(i, v.Struct) }

// Replay_Round_Promise is a wrapper for a Replay_Round promised by a client call.
type Replay_Round_Promise struct{ *capnp.Pipeline }

func (p Replay_Round_Promise) Struct() (Replay_Round, error) {
	s, err := p.Pipeline.Struct()
	return Replay_Round{s}, err
}

func (p Replay_Round_Promise) EndBoard() Board_Promise {
	return Board_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type Faction uint16

// Values of Faction.
const (
	Faction_mine     Faction = 0
	Faction_opponent Faction = 1
)

// String returns the enum's constant name.
func (c Faction) String() string {
	switch c {
	case Faction_mine:
		return "mine"
	case Faction_opponent:
		return "opponent"

	default:
		return ""
	}
}

// FactionFromString returns the enum value with a name,
// or the zero value if there's no such value.
func FactionFromString(c string) Faction {
	switch c {
	case "mine":
		return Faction_mine
	case "opponent":
		return Faction_opponent

	default:
		return 0
	}
}

type Faction_List struct{ capnp.List }

func NewFaction_List(s *capnp.Segment, sz int32) (Faction_List, error) {
	l, err := capnp.NewUInt16List(s, sz)
	return Faction_List{l.List}, err
}

func (l Faction_List) At(i int) Faction {
	ul := capnp.UInt16List{List: l.List}
	return Faction(ul.At(i))
}

func (l Faction_List) Set(i int, v Faction) {
	ul := capnp.UInt16List{List: l.List}
	ul.Set(i, uint16(v))
}

type Turn struct{ capnp.Struct }
type Turn_Which uint16

const (
	Turn_Which_wait         Turn_Which = 0
	Turn_Which_move         Turn_Which = 1
	Turn_Which_attack       Turn_Which = 2
	Turn_Which_selfDestruct Turn_Which = 3
	Turn_Which_guard        Turn_Which = 4
)

func (w Turn_Which) String() string {
	const s = "waitmoveattackselfDestructguard"
	switch w {
	case Turn_Which_wait:
		return s[0:4]
	case Turn_Which_move:
		return s[4:8]
	case Turn_Which_attack:
		return s[8:14]
	case Turn_Which_selfDestruct:
		return s[14:26]
	case Turn_Which_guard:
		return s[26:31]

	}
	return "Turn_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// Turn_TypeID is the unique identifier for the type Turn.
const Turn_TypeID = 0x812bccd38a6bb1d6

func NewTurn(s *capnp.Segment) (Turn, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Turn{st}, err
}

func NewRootTurn(s *capnp.Segment) (Turn, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0})
	return Turn{st}, err
}

func ReadRootTurn(msg *capnp.Message) (Turn, error) {
	root, err := msg.RootPtr()
	return Turn{root.Struct()}, err
}

func (s Turn) String() string {
	str, _ := text.Marshal(0x812bccd38a6bb1d6, s.Struct)
	return str
}

func (s Turn) Which() Turn_Which {
	return Turn_Which(s.Struct.Uint16(0))
}
func (s Turn) Id() uint32 {
	return s.Struct.Uint32(4)
}

func (s Turn) SetId(v uint32) {
	s.Struct.SetUint32(4, v)
}

func (s Turn) SetWait() {
	s.Struct.SetUint16(0, 0)

}

func (s Turn) Move() Direction {
	return Direction(s.Struct.Uint16(2))
}

func (s Turn) SetMove(v Direction) {
	s.Struct.SetUint16(0, 1)
	s.Struct.SetUint16(2, uint16(v))
}

func (s Turn) Attack() Direction {
	return Direction(s.Struct.Uint16(2))
}

func (s Turn) SetAttack(v Direction) {
	s.Struct.SetUint16(0, 2)
	s.Struct.SetUint16(2, uint16(v))
}

func (s Turn) SetSelfDestruct() {
	s.Struct.SetUint16(0, 3)

}

func (s Turn) SetGuard() {
	s.Struct.SetUint16(0, 4)

}

// Turn_List is a list of Turn.
type Turn_List struct{ capnp.List }

// NewTurn creates a new list of Turn.
func NewTurn_List(s *capnp.Segment, sz int32) (Turn_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 0}, sz)
	return Turn_List{l}, err
}

func (s Turn_List) At(i int) Turn { return Turn{s.List.Struct(i)} }

func (s Turn_List) Set(i int, v Turn) error { return s.List.SetStruct(i, v.Struct) }

// Turn_Promise is a wrapper for a Turn promised by a client call.
type Turn_Promise struct{ *capnp.Pipeline }

func (p Turn_Promise) Struct() (Turn, error) {
	s, err := p.Pipeline.Struct()
	return Turn{s}, err
}

type Direction uint16

// Values of Direction.
const (
	Direction_north Direction = 0
	Direction_south Direction = 1
	Direction_east  Direction = 2
	Direction_west  Direction = 3
	Direction_none  Direction = 4
)

// String returns the enum's constant name.
func (c Direction) String() string {
	switch c {
	case Direction_north:
		return "north"
	case Direction_south:
		return "south"
	case Direction_east:
		return "east"
	case Direction_west:
		return "west"
	case Direction_none:
		return "none"

	default:
		return ""
	}
}

// DirectionFromString returns the enum value with a name,
// or the zero value if there's no such value.
func DirectionFromString(c string) Direction {
	switch c {
	case "north":
		return Direction_north
	case "south":
		return Direction_south
	case "east":
		return Direction_east
	case "west":
		return Direction_west
	case "none":
		return Direction_none

	default:
		return 0
	}
}

type Direction_List struct{ capnp.List }

func NewDirection_List(s *capnp.Segment, sz int32) (Direction_List, error) {
	l, err := capnp.NewUInt16List(s, sz)
	return Direction_List{l.List}, err
}

func (l Direction_List) At(i int) Direction {
	ul := capnp.UInt16List{List: l.List}
	return Direction(ul.At(i))
}

func (l Direction_List) Set(i int, v Direction) {
	ul := capnp.UInt16List{List: l.List}
	ul.Set(i, uint16(v))
}

type CellType uint16

// Values of CellType.
const (
	CellType_invalid CellType = 0
	CellType_valid   CellType = 1
	CellType_spawn   CellType = 2
)

// String returns the enum's constant name.
func (c CellType) String() string {
	switch c {
	case CellType_invalid:
		return "invalid"
	case CellType_valid:
		return "valid"
	case CellType_spawn:
		return "spawn"

	default:
		return ""
	}
}

// CellTypeFromString returns the enum value with a name,
// or the zero value if there's no such value.
func CellTypeFromString(c string) CellType {
	switch c {
	case "invalid":
		return CellType_invalid
	case "valid":
		return CellType_valid
	case "spawn":
		return CellType_spawn

	default:
		return 0
	}
}

type CellType_List struct{ capnp.List }

func NewCellType_List(s *capnp.Segment, sz int32) (CellType_List, error) {
	l, err := capnp.NewUInt16List(s, sz)
	return CellType_List{l.List}, err
}

func (l CellType_List) At(i int) CellType {
	ul := capnp.UInt16List{List: l.List}
	return CellType(ul.At(i))
}

func (l CellType_List) Set(i int, v CellType) {
	ul := capnp.UInt16List{List: l.List}
	ul.Set(i, uint16(v))
}

const schema_834c2fcbeb96c6bd = "x\xda\x84Vml\x14U\x17>\xe7\xde\x99\x9dm\xdf" +
	"\xee\xbb\x1df\x9b\x17^-\x1b\x130R\xa5\xd2\"\x91" +
	"4\x9a\x96R\x846\xa0{\x9b\x1ac\xa2?fw\x07" +
	":t;\xb3\xcc\xccR*b\xe5\xc3\xa4\xa8%`\xc0" +
	"\x00\xd2HAB14\xa4\x06\x0c\x8a\xfe0~D\x81" +
	"\xf8\x03\xbfbbL4\xa0T\xc2\x0f\x8d\x10\x13\x941" +
	"wf?\x86\xdd\x8a\x7f\xda\xbbw\xce<\xe7\x9c\xe7<" +
	"\xcf\xbd\xb3\xa0\x83\xb6\x91&\xf1\xe7\x10\x00K\x88!\xf7" +
	"\xeb\xc9\xbe\x17\xbf8\x7f\xeff`\x11D\xf7\xfdO^" +
	"\xbdr\xee\xfe\x95[\xa1C\x12\x01\x94'\xe9+\x8aJ" +
	"%\x80\x85O\xd3'\x10\xd0\x1d\xbex\xe8\xf2\xf0Sw" +
	"\x8f\x80\\\x8b\x00\"\xf2G\x13B5\x02*\xa7\x84V" +
	"@\xf7\xc67\xdb>\xfd\xcf\x95wv\x05\x03.\x09\x84" +
	"\x07Ly\x01\x9d\xbf\xbcwm\xe3\xa5\xdf\xf6\x80\x1c\x09" +
	"\xa4\x13\x89\x04\xa0T\x89?*u\"_\xc9\xe2\x00\xa0" +
	";\xe3\xd0\xb9\x0f\xef<)\xec\x059BK\xb1\x80\xca" +
	":\xf1\xac\xb2\xc9\x0b\x1c\x14\x97+c|\xe5^\xfd\x89" +
	"=\x97\x08\xcf\x1e\x059Bn\x09\xde.\xbe\xa5\xec\xf2" +
	"\x82G\xc4\xc7\x00\xdd\xed;\xae\xbft\xa6i\xe6\xc1\xe9" +
	"*\x18\x13/(\x13^\xec1\xaf\x82\xda\xc5;\xbf\xbb" +
	"#qm\x8c\xb3\x13@\x15xD]h\xbfR\x1f\xe2" +
	"\x0d\xce\x0a\xc599\x1f\x8d\xc7\xa6Nm}\xe60\xc8" +
	"1t\x7f\x18\xb9~>\x9b8=\x99\x07\x9e']P" +
	"\x16I|\xd5$q\xe0\x8b\xf3\xd6^\xfe~\xf6\x96\x13" +
	" \xcfB\xf0\xf0\x16\x8eI]\x08B\xe9\xc5[\xcb\xa3" +
	"^\x03\xd2!e\x8f\xf4?\x00eT:\x01\xe8\x1e\x9d" +
	"=t$ws\xfc\xfct\xad,\x0b\x9fUX\x98\xaf" +
	"V\x85y\xc6\xf8\xebo\x9f\xda\xf89\xfd\xb2\x82\xcc\x89" +
	"\xf0\x16e\xd2\x0b\x9c\x08/W\xbe\xe5+\xf75\xfd\xcc" +
	"\x8e-\x877}U\xa6\x0a\x1f\xf9\x83\xf0~\xe53\x1e" +
	"\xb6\xf0\xe3\xb0\xd7\xb8;\x95L\xee\xfe#\xfbk\x05\xf5" +
	"\x97\xaa\xdeU\xaeV\xf1w\xa6\xaa\x86\x00\xdd\xc7\xf7/" +
	":w\xb4Z\xfe\xbd\"\xb0\xbe\xfaMen5\x0f\xbc" +
	"\xab\xfaA\x98\xef&MG\xcd\xea\x8d)T\xb3F\xb6" +
	"\xa5'g\xa1\x91@d3\xa9\x10A\xd7\x15\x10@\xde" +
	"\xd7\x00\xc0vSd\x07\x09\xd6\x93\x9b.\xc6\x90o\x8f" +
	"\xf2\xed\xbd\x14\xd9\x1b\x04\xeb\xe9_|\x9b\x00\xc8c-" +
	"\x00\xec\x00E6N\xb0\x16cH\x01\xe4#\xff\x07`" +
	"\x07)\xb2\xe3\x04#\xc2\x9fn\x0c\x05\x00\xf9\xd8Z\x00" +
	"6N\x91\x9d$\x18\x11o\xb81\x14\x01\xe4\xc9f\x00" +
	"v\x9c\";M0:\xa0\xea\x0e\x84\xa2\xfd\xe6z\x0d" +
	"\xa3\xa5\xee\x011\x0a\xd8\xaa:\x8e\x9a\xea\xab|@\xf5" +
	"4\x86\x81`\x18\xd0\xb5\xb5\xcc\xea\x0e\xcdv j\xe5" +
	"R\x0e\x84\xe2kr\xaa\x95\x86P\xb1q\xea5\xbeD" +
	"ot\xd4>\xad'g\x19s\xba5;\x97ql\x00" +
	"&P\x01\xc0\xe3 \xc2\xab\x0aSds\x08\xc6\x9d\x9c" +
	"e\xd8\xf8_\xc0\x04E\xac-\x19\x1b\x90o\xde\x069" +
	"\xa1Zj\xbf=-n\x8c`<i\xaaV\x1akK" +
	"\xbe\x01\xc4\xda\x00 \xf1\x00\x97\x9a\x86\xa1\xa5\x9cnm" +
	"]N\xd2l\x87O+\\\xc4\x9b\x97\x04`\xf7Pd" +
	"\x0f\x10\x941?\xa9&N\xff}\x14\xd9b\x82n\xca" +
	"\xd2\xd2\x9a\xe1\xe8 \xa9\x19\x1bkK\xca\xf6\x93QU" +
	"G\xb9$a@\x94+*X\xa2\xe7k0-\x00\x9e" +
	"_\xa0\"@\xf1\xc0\xc1\x82\xe5d\xb9\x1d\x88,JC" +
	")?\xbc\x0d\x13\x88e\x8a[\xaae\xe2\x99\x9e\xc1\xac" +
	"\xc6qj<\x05\xd5\xb7{i\xeb\x9a\x01\x90x\x0c\x0d" +
	"\xe9\xc6z5\xa3\xa7\xe3\xfe_;\xab\x0e\x18e5u" +
	"\x1a\xba\xa3\xab\x99vS\xb50]\xc6\x09\xe7x\x0eE" +
	"\xb6 \xc0\xc9\xfc\xe6<Q\x1d\x01\xe2\x8bf\xf4\xb9\x88" +
	"\xa7\xb4L\xa68\xe8h\xe9\xe0+\x1b\xb4\xdfI\xb7\x99" +
	"\xa4\xa67\x8eX1\xf5&\xce\xfc\x06\x8al\x1bOM" +
	"\xfc\xd4\x9bg\x00\xb0g)\xb2a\x822\xa1\xbem^" +
	"\xe0\x9b\xcfSd/\x13\x94\xa9\xe0\xfbf;\xf7\xd26" +
	"\x8al'AY\x10}\xdb\x8c\xb4\x03\xb0a\x8al7" +
	"\x09*\x1d7\xa0\x04\x04%@\x1c,\xacZ{55" +
	"\xe3\xf4\"\x05\x82\x14ph\xb5\x9art\xd3\xc0h\xe9" +
	"|\xf0\x1dSFf\xb7\x96\xcd\xa8\x83\x8d\xddf\xce\xf8" +
	"G2\xdb\x02d>\xdc\x05\xc0\x1e\xa2\xc8V\x10\x8cs" +
	"\xb3\xde\xc6\x1d\x9a\x91\xe6CJ\x03@%\xe3\xc5:\x84" +
	"r\xa15\xe65Trg9\xfdZV\xca\xa8\x83L" +
	"\xc0\xe0\xfd\x80\xcdq\xdeE\x9a\xd5\x14;X\xc6Im" +
	"\xa3\xc8V\x06:\xe8\xe4\xa4vPd\x09>\x13\xe2\xcf" +
	"d\x15\x8f\\A\x91\xa5\x09\xb6\xaeQ\xfb\xb5\xce4\xd6" +
	"\x00\xc1\x1a\xc0!\xdd\xd7[\xa5[[-\x9e/@@" +
	"\xb1\x982\xd5\xe4\xdd\x9c\xb7\xa3\x9a\xb1}/\xfd\x9b\x97" +
	"\xdb\x03^\xb6\xb5\x94\xa59=&H}\x9aQ,-" +
	"i:\x8f\xaa\xfdZ\xe1w\x19OK\xf4\x80e\x0b\x1f" +
	"\x11X\xf8\xdc\x90\xe5. r\x95\xe4\x16\xce,\x00\x98" +
	"\xce\xb6\xed\xa6J\xadt^\xec\x88\xbe\xd8\x9b\x03b'" +
	"\xf9z7\xb7\x94\xc4\x8e\x05\xad\xb7\xe4\xb5~\x80\xcb:" +
	"\x7fG\xeck.\xdd2\xfc\xde\xe0R\x1fm)]1" +
	"\xf1\x01=\xed\xf4\x06\x94\xad\xaf\xe9u\x8a?-3i" +
	":\x01\xd2\x8b\x9f\x13>\xe9qo((\x00A\x01\xca" +
	"gY\xd6Z\x87n\xb5j\x9eO\xbc\xf6\xbc\x8a\x175" +
	"{G\xd2|\xffH\x9a\xdb\x00\x80T\xae\xe7\xff\x04\xb9" +
	"\xae\x01 n\x98\x96\xd3\x1b\xb7\xcd\x9c\xd3\x1b\xd5T\xdb" +
	"\x89\x0eh\xb6\x135LC+C\x7fDME\x0b\xd8" +
	"a\x0f[n\xf0\xb0\xab\xba\x00\xa2\xfd\xba\xa1\xb9f6" +
	"k\x1a\x9a\xe1\x00\xc0\xdf\x01\x00\x00\xff\xff\xed\xdd\xb1\x07"

func init() {
	schemas.Register(schema_834c2fcbeb96c6bd,
		0x812bccd38a6bb1d6,
		0x8d265c88e8a2e488,
		0x91b9eb0bc884d7fb,
		0x95f2e57bf5bcea49,
		0x9804b41cc3cba212,
		0x9d1e08507e51e6ed,
		0xa01831bb8bf68e89,
		0xa1f5501bdc903810,
		0xa37a83b5e914a8c4,
		0xaf821edee86a29e4,
		0xb1b85070ccf68de1,
		0xcca8fe75a57f1ea7,
		0xd403ce7bb5b69f1f,
		0xd57da3828ebb699b,
		0xf170f8946262e9ff,
		0xf4110aa7cb359a55)
}
