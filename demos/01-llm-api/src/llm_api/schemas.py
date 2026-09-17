from marshmallow import Schema, fields, validate, validates_schema, ValidationError


class MessageSchema(Schema):
    role = fields.String(required=True, validate=validate.OneOf(["system", "user", "assistant"]))
    content = fields.String(required=True, validate=validate.Length(min=1))


class ChatInputSchema(Schema):
    messages = fields.List(fields.Nested(MessageSchema), allow_none=True)
    message = fields.String(allow_none=True)
    mode = fields.String(load_default="text", validate=validate.OneOf(["text", "json"]))
    stream = fields.Boolean(load_default=False)

    @validates_schema
    def validate_input(self, data, **kwargs):
        if (data.get("messages") is None) == (data.get("message") is None):
            raise ValidationError("provide exactly one of messages or message")
