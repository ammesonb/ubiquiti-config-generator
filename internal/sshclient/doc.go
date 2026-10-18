// Package sshclient runs commands on a remote device over a reusable SSH connection.
// Use it to exchange command input and output, with time limits for connecting and running commands.
//
// # Connection lifecycle
//
// Connect once, use the connection for the commands you need, then close it when finished.
// Each SSH session is a separate command channel on a shared connection.
//
//	Dial
//	  |
//	  +-- Run ---------> send input, wait for completion, collect output
//	  |                  command session closes after the call
//	  |
//	  +-- NewSession --> start a process and interact with its streams
//	  |                  close this session when finished with the process
//	  |
//	Client.Close ------> close the connection and all remaining sessions
//
// Supply the device address, including its port, a username, authentication, and a host-key verifier before calling Dial.
// The verifier must check a previously trusted key so the connection authenticates the intended device.
// ConnectTimeout limits connection setup and authentication together, and defaults to 30 seconds when zero.
//
// # Running a command
//
// Use Run when:
// 1. You only need to execute a single command, not a several in sequence.
// 2. You can provide the input up front
// File contents may be provided in input, which are delivered as-is on stdin.
// The command itself is interpreted by the remote shell, so quote any arguments you insert into it.
//
// Run collects stdout and stderr separately, with each call to `Run` having isolated buffers to avoid cross-contamination.
// Each call also starts a new session, so shell variables and working-directory changes do not carry over between calls.
//
// CommandTimeout limits each Run call, including opening its session, with a 30-second default.
// Choose a limit long enough for the operation to finish, especially when uploading a file or applying configuration.
//
// # Keeping a process running
//
// Use NewSession when you need a remote process to stay running while other commands execute, or need to interact with its streams as it runs.
// Callers are required to close sessions on completion when interacting in this manner.
// Note that closing this session does not terminate the connection.
//
// NewSession uses CommandTimeout and its context only while opening the channel, so cancellation after NewSession returns has no effect on the returned session.
// Callers must tie the session's close/cleanup operations to the context themselves, if that is desired.
//
// # Cancellation and failures
//
// The effect of cancellation depends on which operation is in progress:
//
//   - Connecting: Dial fails and closes the connection being established.
//     (It should not be possible for any sessions to have been started yet)
//   - Opening a session: the entire connection closes, interrupting every session sharing it.
//     Reconnect before issuing further commands.
//     (However, commands that are already in progress may still continue running)
//   - Executing a command through Run: only that command's session closes.
//     Other sessions and the shared connection remain available.
//
// Opening a session cannot be canceled independently, which is why canceling it affects other work on the connection.
// Closing an SSH session does not guarantee the remote process has stopped, and it does not undo changes already made.
// After a timeout or connection loss, check the device's state before repeating a command that changes it.
//
// Errors identify the failed operation and retain the underlying cause for inspection with errors.Is or errors.As.
// Command output is returned separately from the printable error, so inspect stdout and stderr when diagnosing a remote failure.
// Output and underlying diagnostics may contain credentials or configuration, so sanitize them before logging.
package sshclient
