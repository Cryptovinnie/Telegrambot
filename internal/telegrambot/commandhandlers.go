// Copyright © 2018 BigOokie
//
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package telegrambot

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/BigOokie/skywire-wing-commander/internal/utils"
	"github.com/BigOokie/skywire-wing-commander/internal/wcconst"
	log "github.com/sirupsen/logrus"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "master"
	dbname   = "skycoinbot"
)

func skycoinCliInput(x string) string {
	path := "/home/josh/go/bin/skycoin-cli"
	input := x //enter argument here to run
	cmd := exec.Command(path, input)

	var out bytes.Buffer
	multi := io.MultiWriter(os.Stdout, &out)
	cmd.Stdout = multi

	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}

	//fmt.Printf(out.String())
	return out.String()

}

func logSendError(from string, err error) {
	log.Errorf("%s - Error: %v", from, err)
}

func getSendModeforContext(ctx *BotContext) string {
	var mode string

	if ctx.IsCallBackQuery() {
		// we cannot "whisper" otherwise this will instruct the
		// bot to talk to itself which is prohibuted. We must "yell"
		mode = "yell"
	} else if ctx.IsUserMessage() {
		mode = "whisper"
	}

	return mode
}

// Handler for help command
func (bot *Bot) handleCommandHelp(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)
	bot.SendGAEvent("BotCommand", command, "Handle"+command)
	err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", fmt.Sprintf(wcconst.MsgHelp, bot.config.Telegram.Admin))
	if err != nil {
		logSendError("Bot.handleCommandHelp", err)
	}
	return err
}

// Handler for about command
func (bot *Bot) handleCommandAbout(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)
	bot.SendGAEvent("BotCommand", command, "Handle"+command)
	err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", wcconst.MsgAbout)
	if err != nil {
		logSendError("Bot.handleCommandAbout", err)
	}
	return err
}

// Handler for showconfig command
func (bot *Bot) handleCommandShowConfig(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)
	bot.SendGAEvent("BotCommand", command+"-asmarkdown", "Handle"+command)
	err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", fmt.Sprintf(wcconst.MsgShowConfig, bot.config.String()))
	if err != nil {
		logSendError("Bot.handleCommandShowConfig (Send):", err)
		log.Debug("Bot.handleCommandShowConfig: Attempting to resend as text.")
		bot.SendGAEvent("BotCommand", command+"-astext", "Handle"+command)
		err = bot.Send(ctx, getSendModeforContext(ctx), "text", fmt.Sprintf(wcconst.MsgShowConfig, bot.config.String()))
		if err != nil {
			logSendError("Bot.handleCommandShowConfig (Resend as Text):", err)
		}
	}
	return err
}

// Handler for uptime command
func (bot *Bot) handleCommandGetUptimeLink(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)

	//https://skywirenc.com/?key_list={node1-id}%2C{node2-id}%2C{node3-id}....etc

	var uptimeURL string

	if bot.skyMgrMonitor.IsRunning() {
		// Add Node Keys as parameters to the URL Query
		uptimeURL = fmt.Sprintf("https://skywirenc.com/?key_list=%s", strings.Join(bot.skyMgrMonitor.GetNodeKeyList(), "%2C"))
		bot.SendGAEvent("BotCommand", command+"-isrunning", "Handle"+command)
	} else {
		uptimeURL = "https://skywirenc.com/"
		bot.SendGAEvent("BotCommand", command+"-notrunning", "Handle"+command)
	}
	msg := fmt.Sprintf("Skywirenc.com (%v Nodes)", bot.skyMgrMonitor.GetConnectedNodeCount())
	log.Debugf("Bot.handleCommandGetUptimeLink: %s", msg)
	log.Debugf("Bot.handleCommandGetUptimeLink: uptimeURL: %s", uptimeURL)

	uptimeURLBtn := tgbotapi.NewInlineKeyboardButtonURL(msg, uptimeURL)
	kbRow := tgbotapi.NewInlineKeyboardRow(uptimeURLBtn)
	kb := tgbotapi.NewInlineKeyboardMarkup(kbRow)

	err := bot.SendReplyInlineKeyboard(ctx, kb, "Check Node uptime here:")
	if err != nil {
		logSendError("Bot.handleCommandGetUptimeLink", err)
	}
	return err
}

// Cryptovinnie Handler for balamce command
func (bot *Bot) handleCommandGetBalanceLink(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)

	walletaddress := "7cpQ7t3PZZXvjTst8G7Uvs7XH4LeM8fBPD" //Change this to get wallet address from @username
	websiteURL := "https://explorer.skycoin.net/api/balance?addrs=" + walletaddress
	log.Debugf("Bot.websiteURL: %s", websiteURL)

	res, _ := http.Get(websiteURL)
	temp, _ := ioutil.ReadAll(res.Body) //JSON Body
	c := bot.config.Coins
	err1 := json.Unmarshal(temp, &c)
	if err1 != nil {
		panic(err1)
	}

	var addressBalance = fmt.Sprintf("%s%d%s%d", "Balance:", c.Confirmed.Coins, "\nCoinHours:", c.Confirmed.Hours) //Convert to string %s, %d for int
	log.Debugf("Bot.getJSONBalance and coin hrs: %s", c.Confirmed.Coins, c.Confirmed.Hours)
	bot.SendGAEvent("BotCommand", command+"-isrunning", "Handle"+command)        //send is running command
	err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", addressBalance) //send message here
	if err != nil {
		logSendError("Bot.handleCommandStart", err)
	}
	return err
}

// Cryptovinnie Handler for Create Address
func (bot *Bot) handleCommandCreateAddressLink(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)
	 
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname) //Connect to psql database.
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	success := "Successfully Connected to DB"
	fmt.Println(success)
	log.Debugf("Handle command: %s ", success)

	sqlStatement := `SELECT id, public_wallet FROM users WHERE telegram_username=$1;`
	var public_wallet string
	var telegram_username string
	var UserName = ctx.message.Chat.UserName
	var chatid = ctx.message.Chat.ID

	row := db.QueryRow(sqlStatement, UserName)
	switch err := row.Scan(&telegram_username, &public_wallet); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")       //User was not found in DB so create address
		Input := skycoinCliInput("addressGenerate") //String to enter after skycoin-cli
		AddrCreated := Input                        //Save created Address to AddrCreated
		//Then Save created wallet to SQL DB
		sqlStatement := `
										INSERT INTO users (chatid, telegram_username, public_wallet, public_address, private_key)
										VALUES ($1, $2, $3, $4, $5)
										RETURNING id`
		id := 0
		err = db.QueryRow(sqlStatement, chatid, UserName, AddrCreated, AddrCreated, AddrCreated).Scan(&id) //Save variables to SQL table
		if err != nil {
			panic(err)
		}
		//User address did not exist created address and added to SQL table. 
		fmt.Println("New record ID is:", id)
		log.Debugf("Bot.NewRecordIDis: %s", id)
		err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", AddrCreated) //send message here
		log.Debugf("Bot.AddressCreatedis: %s", AddrCreated)
		if err != nil {
			logSendError("Bot.handleCommandStart", err)
		}
		return err

	}
	//Address alread created and exists.
	err1 := bot.Send(ctx, getSendModeforContext(ctx), "markdown", public_wallet) //send message here
	log.Debugf("Bot.AddressCreatedis: %s", public_wallet)
	if err1 != nil {
		logSendError("Bot.handleCommandStart", err)
	}
	return err1

}

// Cryptovinnie Handler for send skycommand
func (bot *Bot) handleCommandSendSky(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)
	//1. Check user balance from telegram username 
	//2. Check user @recipient has address in Database 
	//	2a. If not create address for @recipient
	//3. Compare amount to send / Avalilable balance. 
	//4. Create raw transaction 
	//5. Send tx details to User. 

	//1. Check user balance from telegram username use CreateAddress()?
	sqlStatement := `SELECT id, public_wallet FROM users WHERE telegram_username=$1;`
	var public_wallet string
	var telegram_username string
	
	var UserName = ctx.message.Chat.UserName
	var chatid = ctx.message.Chat.ID
	var RecipientUserName := "@Synth"
	var AmounttoSend := 1 //enter in arg from message here.

	row := db.QueryRow(sqlStatement, UserName)
	switch err := row.Scan(&telegram_username, &public_wallet); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")       //User was not found in DB so create address
		Input := skycoinCliInput("addressGenerate") //String to enter after skycoin-cli
		AddrCreated := Input                        //Save created Address to AddrCreated
		//Then Save created wallet to SQL DB
		sqlStatement := `
										INSERT INTO users (chatid, telegram_username, public_wallet, public_address, private_key)
										VALUES ($1, $2, $3, $4, $5)
										RETURNING id`
		id := 0
		err = db.QueryRow(sqlStatement, chatid, UserName, AddrCreated, AddrCreated, AddrCreated).Scan(&id) //Save variables to SQL table
		if err != nil {
			panic(err)
		}
		//User address did not exist created address and added to SQL table. 
		fmt.Println("New record ID is:", id)
		log.Debugf("Bot.NewRecordIDis: %s", id)
		err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", AddrCreated) //send message here
		log.Debugf("Bot.AddressCreatedis: %s", AddrCreated)
		if err != nil {
			logSendError("Bot.handleCommandStart", err)
		}
		return err
		var sendersPublicWallet := AddrCreated 
	}
	//Address alread created and exists.
	var sendersPublicWallet := public_wallet 
	err1 := bot.Send(ctx, getSendModeforContext(ctx), "markdown", public_wallet) //send message here
	log.Debugf("Bot.AddressCreatedis: %s", public_wallet)
	if err1 != nil {
		logSendError("Bot.handleCommandStart", err)
	}
	return err1

	//2. Check user @recipient has address in Database
	row := db.QueryRow(sqlStatement, RecipientUserName)
	switch err := row.Scan(&telegram_username, &public_wallet); err {
		fmt.Println("No rows were returned!")       //User was not found in DB so create address
		Input := skycoinCliInput("addressGenerate") //String to enter after skycoin-cli
		RecipientAddrCreated := Input                        //Save created Address to AddrCreated
		id := 0
		err = db.QueryRow(sqlStatement, chatid, RecipientUserName, RecipientAddrCreated, RecipientAddrCreated, RecipientAddrCreated).Scan(&id) //Save variables to SQL table
		if err != nil {
			panic(err)
		}
		//User address did not exist created address and added to SQL table. 
		fmt.Println("New record ID is:", id)
		log.Debugf("Bot.NewRecordIDis: %s", id)
		err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", RecipientAddrCreated) //send message here Recipient did not have address. 
		log.Debugf("Bot.RecipientAddressCreatedis: %s", RecipientAddrCreated)
		if err != nil {
			logSendError("Bot.handleCommandStart", err)
		}
		return err
		var recipientPublicWallet := RecipientAddrCreated
	}
	//@Recipient has valid address. 
	err1 := bot.Send(ctx, getSendModeforContext(ctx), "markdown", public_wallet) //send message here of address 
	log.Debugf("Bot.RecipientAddressis: %s", public_wallet)
	if err1 != nil {
		logSendError("Bot.handleCommandStart", err)
	}
	return err1
	var recipientPublicWallet := public_wallet
	
	//3. Compare amount to send / Avalilable balance. 
	walletaddress := sendersPublicWallet
	websiteURL := "https://explorer.skycoin.net/api/balance?addrs=" + walletaddress
	log.Debugf("Bot.websiteURL: %s", websiteURL)
	res, _ := http.Get(websiteURL)
	temp, _ := ioutil.ReadAll(res.Body) //JSON Body
	c := bot.config.Coins
	err1 := json.Unmarshal(temp, &c)
	if err1 != nil {
		panic(err1)
	}
	var ConfirmedSkyBalance := c.Confirmed.Coins
	var ConfirmedBalanceHrs := c.Confirmed.Hours
	var addressBalance = fmt.Sprintf("%s%d%s%d", "Balance:", ConfirmedSkyBalance, "\nCoinHours:", ConfirmedBalanceHrs) //Convert to string %s, %d for int
	log.Debugf("Bot.SenderspublicWalletBalance: %s", ConfirmedSkyBalance)
	log.Debugf("Bot.SenderspublicWalletBalanceHours: %s", ConfirmedBalanceHrs)
	//bot.SendGAEvent("BotCommand", command+"-isrunning", "Handle"+command)        //send is running command
	err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", addressBalance) //send message here
	if err != nil {
		logSendError("Bot.handleCommandStart", err)
	}
	return err
	
	//If amount wanting to send is greater then Available balance send error. 
		if AmounttoSend > ConfirmedSkyBalance {
			balance := fmt.Sprintf("%s%d%s%d", "Unable to Send: " , AmounttoSend, "Spendable Amount is: ", ConfirmedSkyBalance) //Convert to string %s, %d for int
			log.Debugf("Bot.Amounttosend: %s", balance)
		}
		else {
		// Amount to spend is less then ConfirmedSkycoinBalance. 
		// Create transaction here and send 
		Input := skycoinCliInput("sendTo -amount "+ AmounttoSend + "-sendto" recipientPublicWallet + "-fromaddress " + sendersPublicWallet + "-changeaddress "+ sendersPublicWallet) //String to enter after skycoin-cli "Sendto"
		SendTransaction := Input                        											//Save transaction Json data. 
		
		ConfirmationTx := fmt.Sprintf("%s%d", "Transaction: " , SendTransaction) //Convert to string %s, %d for int
		err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", ConfirmationTx) //send message here Recipient did not have address. 
		log.Debugf("Bot.ConfirmationTX: %s", ConfirmationTx)
		if err != nil {
			logSendError("Bot.handleCommandStart", err)
		}
		return err
		}

}


}


	walletaddress := "7cpQ7t3PZZXvjTst8G7Uvs7XH4LeM8fBPD" //Change this to get wallet address from @username
	websiteURL := "https://explorer.skycoin.net/api/balance?addrs=" + walletaddress
	log.Debugf("Bot.websiteURL: %s", websiteURL)

	res, _ := http.Get(websiteURL)
	temp, _ := ioutil.ReadAll(res.Body) //JSON Body
	c := bot.config.Coins
	err1 := json.Unmarshal(temp, &c)
	if err1 != nil {
		panic(err1)
	}

	var addressBalance = fmt.Sprintf("%s%d%s%d", "Balance:", c.Confirmed.Coins, "\nCoinHours:", c.Confirmed.Hours) //Convert to string %s, %d for int
	log.Debugf("Bot.getJSONBalance and coin hrs: %s", c.Confirmed.Coins, c.Confirmed.Hours)
	bot.SendGAEvent("BotCommand", command+"-isrunning", "Handle"+command)        //send is running command
	err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", addressBalance) //send message here
	if err != nil {
		logSendError("Bot.handleCommandStart", err)
	}
	return err
}



// Handler for start command
func (bot *Bot) handleCommandStart(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)

	if bot.skyMgrMonitor.IsRunning() {
		log.Debug(wcconst.MsgMonitorAlreadyStarted)
		bot.SendGAEvent("BotCommand", command+"-isrunning", "Handle"+command)
		err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", wcconst.MsgMonitorAlreadyStarted)
		if err != nil {
			logSendError("Bot.handleCommandStart", err)
		}
		return err
	}
	bot.SendGAEvent("BotCommand", command+"-notrunning", "Handle"+command)

	log.Debug(wcconst.MsgMonitorStart)
	cancelContext, cancelFunc := context.WithCancel(context.Background())
	monitorStatusMsgChan := make(chan string)

	// Start the Event Monitor - provide cancelContext
	go bot.monitorEventLoop(cancelContext, ctx, monitorStatusMsgChan)
	// Start monitoring the local Manager - provide cancelContext
	go bot.skyMgrMonitor.RunManagerMonitor(cancelContext, cancelFunc, monitorStatusMsgChan, bot.config.Monitor.IntervalSec)
	// Start monitoring the local Manager - provide cancelContext
	//go bot.skyMgrMonitor.RunDiscoveryMonitor(cancelContext, monitorStatusMsgChan, bot.config.Monitor.DiscoveryMonitorIntMin)

	err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", wcconst.MsgMonitorStart)
	if err != nil {
		logSendError("Bot.handleCommandStart", err)
	}
	return err
}

// Handler for stop command
func (bot *Bot) handleCommandStop(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)

	if bot.skyMgrMonitor.IsRunning() {
		bot.SendGAEvent("BotCommand", command+"-isrunning", "Handle"+command)
		log.Debug(wcconst.MsgMonitorStop)
		bot.skyMgrMonitor.StopManagerMonitor()
		log.Debug(wcconst.MsgMonitorStopped)
		err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", wcconst.MsgMonitorStop)
		if err != nil {
			logSendError("Bot.handleCommandStop", err)
		}
		return err
	}

	bot.SendGAEvent("BotCommand", command+"-notrunning", "Handle"+command)
	log.Debug(wcconst.MsgMonitorNotRunning)
	err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", wcconst.MsgMonitorNotRunning)
	if err != nil {
		logSendError("Bot.handleCommandStop", err)
	}
	return err
}

// Handler for status command
func (bot *Bot) handleCommandStatus(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)

	if !bot.skyMgrMonitor.IsRunning() {
		// Monitor not running
		bot.SendGAEvent("BotCommand", command+"-notrunning", "Handle"+command)
		err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", wcconst.MsgMonitorNotRunning)
		if err != nil {
			logSendError("Bot.handleCommandStatus", err)
		}
		return err
	}

	bot.SendGAEvent("BotCommand", command+"-isrunning", "Handle"+command)
	// Build Status Check Message
	msg := bot.skyMgrMonitor.BuildConnectionStatusMsg(wcconst.MsgStatus)
	err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", msg)
	if err != nil {
		logSendError("Bot.handleCommandStatus", err)
	}
	return err
}

// Handler for help CheckUpdate
func (bot *Bot) handleCommandCheckUpdate(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)
	bot.SendGAEvent("BotCommand", command, "Handle"+command)

	err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", "Checking for updates...")
	if err != nil {
		logSendError("Bot.handleCommandCheckUpdate", err)
		return err
	}

	updateAvailable, updateMsg := utils.UpdateAvailable("BigOokie", "skywire-wing-commander", wcconst.BotVersion)
	if updateAvailable {
		bot.SendGAEvent("BotCommand", command+"-updateavailable", "Handle"+command)
		err = bot.Send(ctx, getSendModeforContext(ctx), "markdown", fmt.Sprintf("*Update available:* %s", updateMsg))
	} else {
		bot.SendGAEvent("BotCommand", command+"-uptodate", "Handle"+command)
		err = bot.Send(ctx, getSendModeforContext(ctx), "markdown", fmt.Sprintf("*Up to date:* %s", updateMsg))
	}

	if err != nil {
		logSendError("Bot.handleCommandCheckUpdate", err)
	}
	return err
}

// Handler for help handleCommandShowMenu
func (bot *Bot) handleCommandShowMenu(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)
	bot.SendGAEvent("BotCommand", command, "Handle"+command)

	err := bot.SendMainMenuMessage(ctx)
	if err != nil {
		logSendError("Bot.handleCommandShowMenu", err)
	}
	return err
}

/*
// Handler for nodes command
func (bot *Bot) handleCommandListNodes(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)

	if bot.skyMgrMonitor.GetConnectedNodeCount() == 0 {
		log.Debug("Bot.handleCommandListNodes: No connected Nodes.")
		err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", "No connected Nodes.")
		if err != nil {
			logSendError("Bot.handleCommandListNodes", err)
		}
		return err
	}

	var replyKeyboard tgbotapi.InlineKeyboardMarkup
	var keyboard [][]tgbotapi.InlineKeyboardButton
	var btnrow []tgbotapi.InlineKeyboardButton
	var btn tgbotapi.InlineKeyboardButton

	// Iterate the connectedNodes and build a keyboard with one button
	// containing the Node Key per row
	for _, v := range bot.skyMgrMonitor.GetNodeKeyList() {
		log.Debugf("Bot.handleCommandListNodes: Creating button for Node: %s", v)
		btn = tgbotapi.NewInlineKeyboardButtonData(v, v)
		btnrow = tgbotapi.NewInlineKeyboardRow(btn)
		keyboard = append(keyboard, btnrow)
	}

	replyKeyboard = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}

	err := bot.SendReplyInlineKeyboard(ctx, replyKeyboard, "Nodes")
	if err != nil {
		log.Error(err)
	}

	return err
}
*/

// Handler for help DoUpdate
func (bot *Bot) handleCommandDoUpdate(ctx *BotContext, command, args string) error {
	log.Debugf("Handle command: %s args: %s", command, args)
	bot.SendGAEvent("BotCommand", command, "Handle"+command)

	err := bot.Send(ctx, getSendModeforContext(ctx), "markdown", "*Initiating update...*")
	if err != nil {
		logSendError("Bot.handleCommandCheckUpdate", err)
		return err
	}

	updateAvailable, updateMsg := utils.UpdateAvailable("BigOokie", "skywire-wing-commander", wcconst.BotVersion)
	if !updateAvailable {
		return bot.Send(ctx, getSendModeforContext(ctx), "markdown", fmt.Sprintf("*Already up to date:* %s", updateMsg))
	}

	err = bot.Send(ctx, getSendModeforContext(ctx), "markdown", fmt.Sprintf("*Update available:* %s", updateMsg))
	if err != nil {
		logSendError("Bot.handleCommandCheckUpdate", err)
		return err
	}

	if utils.DoUpgrade() {
		err = bot.Send(ctx, getSendModeforContext(ctx), "markdown", "Upgrade succeeded.")
	} else {
		err = bot.Send(ctx, getSendModeforContext(ctx), "markdown", "Upgrade failed.")
	}
	if err != nil {
		logSendError("Bot.handleCommandCheckUpdate", err)
		return err
	}

	return nil
}

func (bot *Bot) handleDirectMessageFallback(ctx *BotContext, text string) (bool, error) {
	errmsg := fmt.Sprintf("Sorry, I only take commands. '%s' is not a command.\n\n%s", text, wcconst.MsgHelpShort)
	log.Debugf(errmsg)
	bot.SendGAEvent("BotCommandError", text, "HandleMessageFallback")
	return true, bot.Reply(ctx, "markdown", errmsg)
}

// AddPrivateMessageHandler adds a private MessageHandler to the Bot
func (bot *Bot) AddPrivateMessageHandler(handler MessageHandler) {
	bot.privateMessageHandlers = append(bot.privateMessageHandlers, handler)
}

// AddGroupMessageHandler adds a group MessageHandler to the Bot
func (bot *Bot) AddGroupMessageHandler(handler MessageHandler) {
	bot.groupMessageHandlers = append(bot.groupMessageHandlers, handler)
}

// monitorEventLoop monitors for event messages from the SkyMgrMonitor (when running).
// Its also responsible for managing the Heartbeat (if configured)
func (bot *Bot) monitorEventLoop(runctx context.Context, botctx *BotContext, statusMsgChan <-chan string) {
	tickerHB := time.NewTicker(bot.config.Monitor.HeartbeatIntMin)
	bot.SendGAEvent("BotMonitoring", "Start", "Bot Monitoring Started")
	for {
		select {
		// Monitor Status Message
		case msg := <-statusMsgChan:
			bot.SendGAEvent("BotMonitoring", "ReceiveMonitorStatusMessage", "Receive Monitor Status Message")
			if msg != "" {
				log.Debugf("Bot.monitorEventLoop: Status event: %s", msg)
				err := bot.Send(botctx, getSendModeforContext(botctx), "markdown", msg)
				if err != nil {
					logSendError("Bot.monitorEventLoop", err)
				}
			}

		// Heartbeat ticker event
		case <-tickerHB.C:
			log.Debug("Bot.monitorEventLoop - Heartbeat event")
			bot.SendGAEvent("BotMonitoring", "ReceiveHeartBeat", "Receive Monitor HeartBeat")
			// Build Heartbeat Status Message
			msg := bot.skyMgrMonitor.BuildConnectionStatusMsg(wcconst.MsgHeartbeat)
			log.Debug(msg)
			if msg != "" {
				err := bot.Send(botctx, getSendModeforContext(botctx), "markdown", msg)
				if err != nil {
					logSendError("Bot.handleCommandStatus", err)
				}
			}

		// Context has been cancelled. Shutdown
		case <-runctx.Done():
			log.Debugln("Bot.monitorEventLoop - Done event.")
			bot.SendGAEvent("BotMonitoring", "ReceivedStop", "Receive Monitor Stop")
			return
		}
	}
}
