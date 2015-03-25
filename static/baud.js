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
  textarea.value = TextCanvas.fill(textarea.cols, textarea.rows);
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
// MessageRepo
//

MessageRepo = Baud.MessageRepo = {};

MessageRepo.url = '/api';

MessageRepo.getMessages = function(finished) {
  get(MessageRepo.url, function(status, text) {
    finished(JSON.parse(text));
  });
}

MessageRepo.postMessage = function(name, message, finished) {
  post(MessageRepo.url, { name: name, message: message }, function(status, text) {
    finished(JSON.parse(text));
  });
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
  MessageInput.disable();
  MessageRepo.postMessage(name, message, function(messages) {
    MessageInput.clear();
    MessageInput.enable();
    MessageList.render(messages);
  });
}

MessageInput.clear = function() {
  TextCanvas.clear(MessageInput.messageEl);
  MessageInput.nameEl.value = '';
}

MessageInput.init();

// 
// MessageList
//

var MessageList = Baud.MessageList = {};

MessageList.template = document.getElementById('message-template').innerHTML;

MessageList.container = document.getElementById('message-list');

MessageList.init = function() {
  MessageRepo.getMessages(MessageList.render);
}

MessageList.render = function(messages) {
  var html = '';
  var template = MessageList.template;
  for(var i = 0; i < messages.length; i++) {
    var text = template
      .replace(/__NAME__/g, messages[i].name)
      .replace(/__MESSAGE__/g, messages[i].message)
      .replace(/__DATE__/g, MessageList.date(messages[i].date))
    html += text;
  }
  MessageList.container.innerHTML = html;
}

MessageList.date = function(timestamp) {
  var now = Math.floor((new Date()).getTime() / 1000);
  return (now - timestamp) + ' seconds ago';
}

MessageList.init();
