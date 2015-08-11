// Interface to other players on the net.

// An implementation of this interface will initialize a runtime
// capable of making secure RPCs, then begin to accept and send RPCs
// on behalf of the client of the interface.

// The implementation will be held by a game table?

package game

type Manager interface {
	// Returns a channel on which new players may appear.
	Players() chan Player

	// Send a ball (associated with a particular player, though this
	// won't matter for volleyball) to some other player.  The target
	// player is decided by the table, since the table knows how the ball is moving
	// and how the players are arranged.
	SendBall(*Ball, *Player) error

	// Tell all players that you are leaving, and won't be sending or
	// accepting any more data.
	Quit() error
}
