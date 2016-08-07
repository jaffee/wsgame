var stage = new PIXI.Container(),
    renderer = PIXI.autoDetectRenderer(256, 256);

document.body.appendChild(renderer.view);

PIXI.loader
    .add("images/red.png")
    .add("images/blue.png")
    .load(setup);

var csprite;
var bgsprite;

var ws;
var open = false;
var dudes = {};

function setup() {
    ws = new WebSocket("ws://localhost:12345/game");
    ws.onmessage = function (event) {
        // console.log(event.data);
        data = JSON.parse(event.data);
        for (i = 0; i < data.length; i++){
            node = data[i];
            if (dudes[node.ID] == undefined) {
                dudes[node.ID] = new PIXI.Sprite(
                    PIXI.loader.resources[node.Texture].texture
                );
                console.log("new dude", dudes[node.ID]);
                stage.addChild(dudes[node.ID]);
            }
            sp = dudes[node.ID];
            if (node.X != sp.x) {
                console.log("changing: from", sp.x, " ", node.X);
            }
            sp.x = node.X; sp.y = node.Y;
            sp.width = node.Width;
            sp.height = node.Height;
        }
    };

    ws.onopen = function () {
        open = true;
    };
    // ws.onerror = function () {
    //     open = false;
    // };
    ws.onclose = function () {
        open = false;
    };
    gameLoop();
}

var sendIfOpen = function(astr) {
    console.log("Sending: " + astr);
    if (open) {
        ws.send(astr);
    } else {
        console.log("Not open!");
    }
};

function keyboard(keyCode) {
    var key = {};
    key.code = keyCode;
    key.isDown = false;
    key.isUp = true;
    key.press = undefined;
    key.release = undefined;
    //The `downHandler`
    key.downHandler = function(event) {
        if (event.keyCode === key.code) {
            if (key.isUp && key.press) key.press();
            key.isDown = true;
            key.isUp = false;
        }
        event.preventDefault();
    };

    //The `upHandler`
    key.upHandler = function(event) {
        if (event.keyCode === key.code) {
            if (key.isDown && key.release) key.release();
            key.isDown = false;
            key.isUp = true;
        }
        event.preventDefault();
    };

    //Attach event listeners
    window.addEventListener(
        "keydown", key.downHandler.bind(key), false
    );
    window.addEventListener(
        "keyup", key.upHandler.bind(key), false
    );
    return key;
};

var left = keyboard(37),
    up = keyboard(38),
    right = keyboard(39),
    down = keyboard(40);

up.press = function () { console.log('up'); sendIfOpen('{"Com": 1}');};
down.press = function () { console.log('down'); sendIfOpen('{"Com": 2}');};
right.press = function () { console.log('right'); sendIfOpen('{"Com": 3}');};
left.press = function () { console.log('left'); sendIfOpen('{"Com": 4}');};


function gameLoop() {
    //Loop this function at 60 frames per second
    requestAnimationFrame(gameLoop);

    //Render the stage to see the animation
    renderer.render(stage);
};

//Start the game loop
