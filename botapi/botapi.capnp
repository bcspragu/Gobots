using Go = import "../../../../zombiezen.com/go/capnproto2/go.capnp";

@0x834c2fcbeb96c6bd;

$Go.package("botapi");
$Go.import("github.com/bcspragu/Gobots/botapi");

using RobotId = UInt32;

interface AiConnector {
  # Bootstrap interface for the server.
  connect @0 ConnectRequest -> ();
}

struct ConnectRequest {
  credentials @0 :Credentials;
  ai @1 :Ai;
}

struct Credentials {
  secretToken @0 :Text;
  botName @1 :Text;
}

interface Ai {
  # Interface that a competitor implements.
  takeTurn @0 (board :InitialBoard) -> (turns :List(Turn));
}

struct Board {
  gameId @4 :Text;

  width @0 :UInt16;
  height @1 :UInt16;
  robots @2 :List(Robot);

  round @3 :Int32;
}

struct InitialBoard {
  board @0 :Board;
  cells @1 :List(CellType);
}

struct Robot {
  id @0 :RobotId;
  x @1 :UInt16;
  y @2 :UInt16;
  health @3 :Int16;
  faction @4 :Faction;
}

struct Replay {
  gameId @0 :Text;
  initial @1 :InitialBoard;
  rounds @2 :List(Round);
  
  struct Round {
    moves @0 :List(Turn);
    endBoard @1 :Board;
    # The board at the end of the round, after applying moves
  }
}

enum Faction {
  mine @0;
  opponent @1;
}

struct Turn {
  id @3 :RobotId;

  union {
    wait @0 :Void;
    # Skip turn; do nothing.

    move @1 :Direction;

    attack @2 :Direction;

    selfDestruct @4 :Void;
    # Does damage to all surrounding bots (even diagonals).

    guard @5 :Void;
  }
}

enum Direction {
  north @0;
  south @1;
  east @2;
  west @3;
  none @4;
}

enum CellType {
  invalid @0;
  valid @1;
  spawn @2;
}
