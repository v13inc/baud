import json
import time
import redis
from flask import Flask, request, send_file, render_template, jsonify
from flask.ext.socketio import SocketIO, emit

settings = {
    'redis': {
        'host': 'localhost',
        'port': 6379,
    },
    'messages': {
        'list_key': 'messages',
        'per_page': 100
    }
}

app = Flask(__name__)
app.config['DEBUG'] = True
socketio = SocketIO(app)

class Message():
    def __init__(self, message, name = 'Anonymous', date = None):
        self.name = name
        self.date = date or round(time.time())
        self.message = message
    
    def data(self):
        return {
            'name': self.name,
            'date': self.date,
            'message': self.message
        }

    def serialize(self):
        return json.dumps(self.data())

    def __repr__(self):
        return self.data()

class MessageRepo():
    connection = None

    @staticmethod
    def conn():
        if MessageRepo.connection:
            return MessageRepo.connection

        MessageRepo.connection = redis.Redis(**settings['redis'])
        return MessageRepo.connection

    @staticmethod
    def get_messages(start = 0, amount = settings['messages']['per_page'], key = settings['messages']['list_key']):
        c = MessageRepo.conn()
        messages = c.zrevrange(key, start, int(start) + int(amount) - 1)
        return [json.loads(m) for m in messages]

    @staticmethod
    def add_message(message, key = settings['messages']['list_key']):
        c = MessageRepo.conn()
        return c.zadd(key, message.serialize(), message.date)

    @staticmethod
    def new_message(message, name):
        m = Message(message, name)
        MessageRepo.add_message(m)
        return m

@socketio.on('send_message')
def handle_message(data):
    message = MessageRepo.new_message(data['message'], data['name'])
    emit('new_message', message.data(), broadcast = True)

# add main route
@app.route('/')
def home():
    return render_template('index.html', messages_json = json.dumps(MessageRepo.get_messages()))

@app.route('/api/messages', methods=['POST'])
def post_message():
    data = validate_message(request.get_json(force = True))
    handle_message(data)
    return json.dumps(data)

@app.route('/api/messages', methods=['GET'])
def get_messages():
    start = request.args.get('start', 0)
    amount = request.args.get('amount', settings['messages']['per_page'])
    return json.dumps(MessageRepo.get_messages(start, amount))

if __name__ == '__main__':
    socketio.run(app)
    #app.run(debug = True)
