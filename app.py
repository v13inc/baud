import json
import time
import redis
from flask import Flask, request, send_file
from flask_restful import Resource, Api

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
api = Api(app)

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
    def get_messages(start = 0, end = settings['messages']['per_page'], key = settings['messages']['list_key']):
        c = MessageRepo.conn()
        messages = c.zrevrange(key, start, end)
        return [json.loads(m) for m in messages]

    @staticmethod
    def add_message(message, key = settings['messages']['list_key']):
        c = MessageRepo.conn()
        return c.zadd(key, message.serialize(), message.date)

    @staticmethod
    def new_message(message, name):
        m = Message(message, name)
        return MessageRepo.add_message(m)

class MessagesResource(Resource):
    def get(self):
        return MessageRepo.get_messages()

    def post(self):
        data = request.form
        MessageRepo.new_message(data['message'], data['name'])
        return MessagesResource.get(self)

api.add_resource(MessagesResource, '/api')

# add main route
@app.route('/')
def home():
    return send_file('index.html')

@app.route('/banners.js')
def banners():
    return send_file('banners.js')

if __name__ == '__main__':
    app.run(debug = True)
