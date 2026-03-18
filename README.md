# Reaper OSC Action (But for Linux)

This is a Elgato StreamDeck Plugin to send Command IDs to the DAW Reaper via OSC. It is written in Go.

Demo Video can be found here: https://www.youtube.com/watch?v=DTwFpP6xsbU

## Installation

### Linux (applicable to this fork)

`git clone https://github.com/lanabretterebner/reaper_osc_action_linux`

`cd reaper_osc_action_linux`

`make`

then copy the folder named "org.smyck.reaper-osc-action.sdPlugin" into your opendeck plugin folder and restart opendeck.

#### OR
go to releases and just download the folder mentioned before!


## Building

If you want to build the plugin yourself you should only need to have a working
Go installation and then run `make`

This will also do the cross compilation and create a universal binary.

After this you can symlink the plugin folder to the StreamDeck plugins folder

`ln -s ./org.smyck.reaper_osc_action.sdPlugin ~/Library/Application\ Support/com.elgato.StreamDeck/Plugins/`

Restart the StreamDeck app and verify in the Preferences / Plugins tab that
the "Reaper OSC Action" plugin appears.

Then place it on one of the StreamDeck Buttons and add "127.0.0.1" as IO, the
port that was configured in Reaper and a command id of your choice.

You can also build a .streamDeckPlugin file by running:
`make plugin`

You need the cli tools `fd` and `streamdeck`

* https://github.com/sharkdp/fd
* https://github.com/elgatosf/cli

## Setup of Reaper

To make this all work:

* Go to Reapers Preferences > Control / OSC / Web
* Click "Add" and to add an OSC control surface
* In the Control Surface Settings dialogue, set the mode to "Configure device IP + local port
* Choose a "Local listen port" of your choice and use that in the Streamdeck Plugin
* The "Listen" button will open a window to check whether Reaper is receiving the messages from StreamDeck
