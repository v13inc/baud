var Baud = window.Baud = {};

// 
// utils
//

var request = Baud.request = function(type, url, data, finished) {
  var req = new XMLHttpRequest();
  req.open(type, url, true);
  req.onreadystatechange = function() {
    console.log('state', req.readyState, req.status);
    if(req.readyState == 4) {
      console.log(req);
      finished(req.status, req.responseText);
    }
  }
  req.send(data);
}

var get = Baud.get = function(url, finished) {
  request('GET', url, null, finished);
}

var post = Baud.post = function(url, data, finished) {
  request('POST', url, JSON.stringify(data), finished);
}

// 
// TextCanvas
//

var TextCanvas = Baud.TextCanvas = {};

TextCanvas.nbsp = '\u00A0'; // this is a super-special non-breaking international space
TextCanvas.blank = '\u00A0'; // this is a super-special non-breaking international space
//TextCanvas.blank = '.';

TextCanvas.init = function(textarea, onSubmit) {
  TextCanvas.clear(textarea);
  textarea.addEventListener('keydown', TextCanvas.keydown.bind(null, textarea, onSubmit));
}

TextCanvas.clear = function(textarea) {
  var cursorPos = textarea.selectionStart;
  textarea.value = TextCanvas.fill(textarea.cols, textarea.rows);
  textarea.selectionStart = textarea.selectionEnd = cursorPos;
}

TextCanvas.chars = function(char, amount) {
  for(var i = 0, text = ''; i < amount; i++) text += char;
  return text;
}

TextCanvas.fill = function(cols, rows) {
  var rows = rows || 1;
  for(var i = 0, text = ''; i < rows; i++) {
    text += TextCanvas.chars(TextCanvas.blank, cols);
  }
  return text;
}

TextCanvas.keydown = function(textarea, onSubmit, e) {
  if(e.keyCode == 13) {
    e.preventDefault();
    onSubmit(TextCanvas.convertText(textarea.value, textarea.cols));
  } else {
    setTimeout(TextCanvas.overwrite.bind(null, textarea), 0);
  }
}

TextCanvas.overwrite = function(textarea) {
  var cursorPos = textarea.selectionStart;
  var text = textarea.value;
  var size = textarea.cols * textarea.rows;
  var diff = text.length - size;
  var start = text.substring(0, Math.min(cursorPos, size));
  var end = text.substring(cursorPos);
  if(diff > 0) { 
    text = start + end.substring(diff);
  } else if(diff < 0) {
    text = start + TextCanvas.fill(-diff) + end;
  }
  if(diff != 0) {
    textarea.value = text
      .replace(/\n/g, TextCanvas.blank)
      .replace(/ /g, TextCanvas.blank);
    textarea.selectionStart = textarea.selectionEnd = cursorPos;
  }
}

TextCanvas.convertText = function(inputText, cols) {
  var text = '';
  for(var i = 0; i < inputText.length; i++) {
    if(i % cols == 0 && i != 0) text += '\n';
    text += inputText[i] == TextCanvas.blank ? ' ' : inputText[i];
  }
  return text;
}

// 
// MessageDispatcher
//

MessageDispatcher = Baud.MessageDispatcher = {};

MessageDispatcher.connected = false;

MessageDispatcher.socket = function() {
  if(MessageDispatcher.conn) return MessageDispatcher.conn;
  var conn = MessageDispatcher.conn = io.connect();
  conn.on('connect', MessageDispatcher.onConnect);
  return conn;
}

MessageDispatcher.onConnect = function() {
  var socket = MessageDispatcher.socket();
  MessageDispatcher.connected = true;
  socket.on('new_message', MessageDispatcher.onMessage);
}

MessageDispatcher.onMessage = function(message) {
  console.log('message', message);
  MessageList.addMessage(message);
}

MessageDispatcher.init = function() {
  MessageDispatcher.socket();
}

MessageDispatcher.sendMessage = function(name, message) {
  MessageDispatcher.socket().emit('send_message', { name: name, message: message });
}

// 
// MessageInput
//

MessageInput = Baud.MessageInput = {};

MessageInput.defaultName = 'Anonymous';

MessageInput.messageEl = document.getElementById('message');

MessageInput.nameEl = document.getElementById('name');

MessageInput.init = function() {
  TextCanvas.init(MessageInput.messageEl, MessageInput.submit);
}

MessageInput.enable = function() {
  MessageInput.messageEl.disabled = false;
  MessageInput.nameEl.disabled = false;
}

MessageInput.disable = function() {
  MessageInput.messageEl.disabled = true;
  MessageInput.nameEl.disabled = true;
}

MessageInput.submit = function(message) {
  var name = MessageInput.nameEl.value.trim() || MessageInput.defaultName;
  MessageDispatcher.sendMessage(name, message);
  MessageInput.clear();
}

MessageInput.clear = function() {
  TextCanvas.clear(MessageInput.messageEl);
  //MessageInput.nameEl.value = '';
}

MessageInput.init();

// 
// MessageList
//

var MessageList = Baud.MessageList = {};

MessageList.template = document.getElementById('message-template').innerHTML;

MessageList.container = document.getElementById('message-list');

MessageList.messages = [];

MessageList.init = function() {
  MessageList.render(MESSAGES);
}

MessageList.render = function(messages) {
  var html = '';
  var template = MessageList.template;
  for(var i = 0; i < messages.length; i++) {
    html += MessageList.renderMessage(messages[i]);
  }
  MessageList.container.innerHTML = html;
  MessageList.messages = messages;
}

MessageList.renderMessage = function(message) {
    return MessageList.template
      .replace(/__NAME__/g, message.name)
      .replace(/__MESSAGE__/g, message.message)
      .replace(/__DATE__/g, MessageList.date(message.date))
}

MessageList.addMessage = function(message) {
  MessageList.messages.splice(0, 0, message);
  MessageList.render(MessageList.messages);
}

MessageList.date = function(timestamp) {
  var now = Math.floor((new Date()).getTime() / 1000);
  var diff = now - timestamp;
  return (diff > 0 ? diff : 0) + ' seconds ago';
}

MessageList.init();
