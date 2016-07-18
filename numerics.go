// showdown2irc - use Showdown chat with an IRC client
// Copyright (C) 2016 Konrad Borowski
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

// IRCNumeric represents reply codes used by IRC protocol
type IRCNumeric int

const (
	// ErrNoSuchNick states that nick doesn't exist.
	ErrNoSuchNick IRCNumeric = 401

	// ErrNoSuchServer states that server doesn't exist.
	ErrNoSuchServer = 402

	// ErrNoSuchChannel states that channel/room doesn't exist.
	ErrNoSuchChannel = 403

	// ErrCannotSendToChan informs an user about failure to send
	// a message.
	//
	// There are many situations when this can happen. In IRC, the
	// message is sent when the user is banned, channel has mode +n and
	// user is not in a channel, channel is moderated (+m) and user is
	// not a voice (+v).
	//
	// On Showdown however, this can happen when moderated chat is
	// enabled and user doesn't have enough privileges, the user is
	// banned, or automatic chat filter prevented the message from being
	// submitted.
	ErrCannotSendToChan = 404

	// ErrTooManyChannels states that user joined too many channels.
	//
	// There is no channel limitation on Showdown
	ErrTooManyChannels = 405

	// ErrWasNoSuchNick sent by WHOWAS says that the user is unknown.
	ErrWasNoSuchNick = 406

	// ErrTooManyTargets says that a given string matched too many
	// targets.
	//
	// Too many targets depends on a particular command. Operators
	// are allowed to use wildcards and host masks as an argument of
	// PRIVMSG. This isn't a concern for showdown2irc, because this
	// program doesn't support existence of operator status.
	//
	// JOIN command is allowed to return ErrTooManyTargets when in a
	// specific situation. The specification calls this situation to
	// be "joining a safe channel using the shortname when there are
	// than one such channel". I don't know what that means, but this
	// program doesn't support this.
	ErrTooManyTargets = 407

	// ErrNoOrigin says that PING or PONG lack the originator
	// parameter.
	ErrNoOrigin = 409

	// ErrNoRecipient says that recipient parameter was omitted in a
	// private message.
	ErrNoRecipient = 411

	// ErrNoTextToSend says that there was no text specified to send.
	ErrNoTextToSend = 412

	// ErrNoTopLevel is caused by trying to use wildcard PRIVMSG on a
	// domain without top level domain part.
	ErrNoTopLevel = 413

	// ErrWildTopLevel is caused by trying to use wildcard PRIVMSG on a
	// domain for which top level domain part is a wildcard (*).
	ErrWildTopLevel = 414

	// ErrUnknownCommand is caused by unrecognized command.
	ErrUnknownCommand = 421

	// ErrNoMOTD says that MOTD file is missing.
	ErrNoMOTD = 422

	// ErrNoAdminInfo says that administrator information is not
	// available.
	ErrNoAdminInfo = 423

	// ErrFileError is a generic file error message.
	ErrFileError = 424

	// ErrNoNicknameGiven is an error caused by not specifying nicknames.
	ErrNoNicknameGiven = 431

	// ErrErroneusNickname is an error caused by having a username
	// that doesn't follow IRC rules for nicknames.
	//
	// Because that's the only nickname that is about invalid usernames
	// (other than ErrNoNicknameGiven which is explicitly about empty
	// usernames), showdown2irc uses it for purposes of marking
	// usernames that were explicitly rejected by server, such as
	// usernames starting with "Guest".
	ErrErroneusNickname = 432

	// ErrNicknameInUse says that the nickname is already used.
	ErrNicknameInUse = 433

	// ErrNickCollision informs about nickname collision between
	// multiple servers.
	ErrNickCollision = 436

	// ErrUserNotInChannel says that an user is not in a channel.
	ErrUserNotInChannel = 441

	// ErrNotOnChannel is caused by trying to use channel affecting
	// command while not on server.
	ErrNotOnChannel = 442

	// ErrUserOnChannel is caused by inviting an user into a room in
	// which an user is already in.
	ErrUserOnChannel = 443

	// ErrNoLogin is message from SUMMON that the administrator is not
	// logged in.
	ErrNoLogin = 444

	// ErrSummonDisabled is a message from SUMMON saying that the
	// command is disabled.
	ErrSummonDisabled = 445

	// ErrUsersDisabled says that USERS is disabled.
	ErrUsersDisabled = 446

	// ErrNotRegistered says that the user is not logged in.
	//
	// This message is caused by trying to use an command before
	// properly authenticating on a server.
	ErrNotRegistered = 451

	// ErrNeedMoreParams is a message saying that a command needs more
	// parameters.
	ErrNeedMoreParams = 461

	// ErrAlreadyRegistered is caused by trying to log in when already
	// logged in.
	ErrAlreadyRegistered = 462

	// ErrNoPermForHost is a message saying that the server is not
	// configured to allow connections from current host.
	ErrNoPermForHost = 463

	// ErrPasswdMismatch says the password for an user is incorrect.
	//
	// This can be also triggered by not specifying a password.
	ErrPasswdMismatch = 464

	// ErrYouAreBannedCreep is caused by trying to connect to a server
	// form which an user is banned.
	//
	// In Showdown terms, this means /globalban.
	ErrYouAreBannedCreep = 465

	// ErrKeySet says that channel key was already set.
	//
	// Showdown doesn't support channel keys.
	ErrKeySet = 467

	// ErrChannelIsFull says that channel exceeded its limits.
	//
	// There are no user limits on Showdown.
	ErrChannelIsFull = 471

	// ErrUnknownMode says that the server doesn't recognize a mode.
	//
	// On Showdown, this is also used for modes that technically are
	// part of IRC RFC, but aren't possible on Showdown.
	ErrUnknownMode = 472

	// ErrInviteOnlyChan says the joined channel is invite only.
	//
	// This isn't used on Showdown, as failure to join a channel
	// because it's invite only looks exactly like failure to enter
	// a channel because it doesn't exist.
	ErrInviteOnlyChan = 473

	// ErrBannedFromChan is caused by trying to join a channel from
	// which an user is banned.
	ErrBannedFromChan = 474

	// ErrBadChannelKey is caused by specifying wrong channel key.
	//
	// There are no channel keys on Showdown.
	ErrBadChannelKey = 475

	// ErrNoPrivileges is shown when command requires to be an IRC
	// operator.
	//
	// As showdown2irc doesn't support IRC operator status, all IRC
	// operator only commands return this.
	ErrNoPrivileges = 481

	// ErrChanOpPrivIsNeeded is shown when a command that requires
	// operator privileges is used, and user doesn't have these.
	//
	// Showdown has different privileges levels with different
	// permissions. For example, while an user may be allowed to mute
	// users, he may not be allowed to roomban. This is also used
	// in those cases.
	ErrChanOpPrivIsNeeded = 482

	// ErrCannotKillServer is caused by lack of permissions to run
	// /kill command.
	ErrCannotKillServer = 483

	// ErrNoOperHost is caused by trying to claim IRC operator status,
	// despite no permissions to do it.
	ErrNoOperHost = 491

	// ErrUmodeUnknownFlag is caused by trying to use an unknown user
	// mode flag.
	ErrUmodeUnknownFlag = 501

	// ErrUsersDoNotMatch is caused by trying to change flags of other
	// users.
	ErrUsersDoNotMatch = 502

	RplWelcome       = 1
	RplYourHost      = 2
	RplCreated       = 3
	RplMyInfo        = 4
	RplBounce        = 5
	RplUserhost      = 302
	RplIson          = 303
	RplAway          = 301
	RplUnaway        = 305
	RplNowAway       = 306
	RplWhoisUser     = 311
	RplWhoisServer   = 312
	RplWhoisOperator = 313
	RplWhoisIdle     = 317
	RplEndOfWhois    = 318
	RplWhoisChannels = 319
	RplWhowasUser    = 314
	RplEndOfWhowas   = 369
	RplListStart     = 321
	RplList          = 322
	RplListEnd       = 323
	RplChannelModeIs = 324
	RplNoTopic       = 331
	RplTopic         = 332
	RplInviting      = 341
	RplSummoning     = 342
	RplVersion       = 351
	RplWhoReply      = 352
	RplEndOfWho      = 315
	RplNamesReply    = 353
	RplEndOfNames    = 366
	RplLinks         = 364
	RplEndOfLinks    = 365
	RplBanList       = 367
	RplEndOfBanList  = 368
	RplInfo          = 371
	RplEndOfInfo     = 374
	RplMOTDStart     = 375
	RplMOTD          = 372
	RplEndOfMOTD     = 376
	RplYouAreOper    = 381
	RplRehashing     = 382
	RplTime          = 391
	RplUsersStart    = 392
	RplUsers         = 393
	RplEndOfUsers    = 394
	RplNoUsers       = 395

	RplTraceLink       = 200
	RplTraceConnecting = 201
	RplTraceHandshake  = 202
	RplTraceUnknown    = 203
	RplTraceOperator   = 204
	RplTraceUser       = 205
	RplTraceServer     = 206
	RplTraceNewType    = 208
	RplTraceLog        = 261

	RplStatsLinkInfo = 211
	RplStatsCommands = 212
	RplEndOfStats    = 219
	RplStatsUptime   = 242
	RplStatsOLine    = 243
	RplUmodeIs       = 221
	RplLuserClient   = 251
	RplLuserOp       = 252
	RplLuserUnknown  = 253
	RplLuserChannels = 254
	RplLuserMe       = 255
	RplAdminMe       = 256
	RplAdminLoc1     = 257
	RplAdminLoc2     = 258
	RplAdminEmail    = 259
)

var numericMessages = map[IRCNumeric]string{
	ErrNoSuchNick:         "%s :No such nick/channel",
	ErrNoSuchServer:       "%s :No such server",
	ErrNoSuchChannel:      "%s :No such channel",
	ErrCannotSendToChan:   "%s :Cannot send to channel",
	ErrTooManyChannels:    "%s :You have joined too many channels",
	ErrWasNoSuchNick:      "%s :There was no such nickname",
	ErrTooManyTargets:     "%s :Duplicate recipients. No message delivered",
	ErrNoOrigin:           ":No origin specified",
	ErrNoRecipient:        ":No recipient given (%s)",
	ErrNoTextToSend:       ":No text to send",
	ErrNoTopLevel:         "%s :No toplevel domain specified",
	ErrWildTopLevel:       "%s :Wildcard in toplevel domain",
	ErrUnknownCommand:     "%s :Unknown command",
	ErrNoMOTD:             ":MOTD File is missing",
	ErrNoAdminInfo:        "%s :No administrative info available",
	ErrFileError:          ":File error doing %s on %s",
	ErrNoNicknameGiven:    ":No nickname given",
	ErrErroneusNickname:   "%s :Erroneus nickname",
	ErrNicknameInUse:      "%s :Nickname is already in use",
	ErrNickCollision:      "%s :Nickname collision KILL",
	ErrUserNotInChannel:   "%s %s :They aren't on that channel",
	ErrNotOnChannel:       "%s :You're not on that channel",
	ErrUserOnChannel:      "%s %s :is already on channel",
	ErrNoLogin:            "%s :User not logged in",
	ErrSummonDisabled:     ":SUMMON has been disabled",
	ErrUsersDisabled:      ":USERS has been disabled",
	ErrNotRegistered:      ":You have not registered",
	ErrNeedMoreParams:     "%s :Not enough parameters",
	ErrAlreadyRegistered:  ":You may not reregister",
	ErrNoPermForHost:      ":Your host isn't among the privileged",
	ErrPasswdMismatch:     ":Password incorrect",
	ErrYouAreBannedCreep:  ":You are banned from this server",
	ErrKeySet:             "%s :Channel key already set",
	ErrChannelIsFull:      "%s :Cannot join channel (+l)",
	ErrUnknownMode:        "%s :is unknown mode char to me",
	ErrInviteOnlyChan:     "%s :Cannot join channel (+i)",
	ErrBannedFromChan:     "%s :Cannot join channel (+b)",
	ErrBadChannelKey:      "%s :Cannot join channel (+k)",
	ErrNoPrivileges:       ":Permission Denied- You're not an IRC operator",
	ErrChanOpPrivIsNeeded: "%s :You're not channel operator",
	ErrCannotKillServer:   ":You cant kill a server!",
	ErrNoOperHost:         ":No O-lines for your host",
	ErrUmodeUnknownFlag:   ":Unknown MODE flag",
	ErrUsersDoNotMatch:    ":Cant change mode for other users",

	RplWelcome:       ":%s",
	RplBounce:        "%s",
	RplUserhost:      ":%s",
	RplIson:          ":%s",
	RplAway:          "%s :%s",
	RplUnaway:        ":You are no longer marked as being away",
	RplNowAway:       ":You have been marked as being away",
	RplWhoisUser:     "%s %s %s * :%s",
	RplWhoisServer:   "%s %s :%s",
	RplWhoisOperator: "%s :is an IRC operator",
	RplWhoisIdle:     "%s %d :seconds idle",
	RplEndOfWhois:    "%s :End of /WHOIS list",
	RplWhoisChannels: "%s: %s",
	RplWhowasUser:    "%s %s %s * :%s",
	RplEndOfWhowas:   "%s :End of WHOWAS",
	RplListStart:     "Channel :Users  Name",
	RplList:          "%s %d :%s",
	RplListEnd:       ":End of /LIST",
	RplChannelModeIs: "%s %s %s",
	RplNoTopic:       "%s: No topic is set",
	RplTopic:         "%s: %s",
	RplInviting:      "%s %s",
	RplSummoning:     "%s :Summoning user to IRC",
	RplVersion:       "%s.%s %s :%s",
	RplWhoReply:      "%s %s %s %s %s %c%s :%d %s",
	RplEndOfWho:      "%s :End of /WHO list",
	RplNamesReply:    "%c %s :%s",
	RplEndOfNames:    "%s :End of /NAMES list",
	RplLinks:         "%s %s :%d %s",
	RplEndOfLinks:    "%s :End of /LINKS list",
	RplBanList:       "%s %s",
	RplEndOfBanList:  "%s :End of channel ban list",
	RplInfo:          ":%s",
	RplEndOfInfo:     ":End of /INFO list",
	RplMOTDStart:     ":- %s Message of the day - ",
	RplMOTD:          ":- %s",
	RplEndOfMOTD:     ":End of /MOTD command",
	RplYouAreOper:    ":You are now an IRC operator",
	RplRehashing:     "%s :Rehashing",
	RplTime:          "%s :%s",
	RplUsersStart:    ":UserID   Terminal  Host",
	RplUsers:         ":%-8s %-9s %-8s",
	RplEndOfUsers:    ":End of users",
	RplNoUsers:       ":Nobody logged in",
}
