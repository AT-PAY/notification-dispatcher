-- +goose Up
-- +goose StatementBegin

-- PAYMENT_SUCCESS Templates
-- =====================================================

-- WebSocket - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('PAYMENT_SUCCESS',
        'WEB_SOCKET',
        'vi',
        'Giao dịch thành công',
        'Bạn đã chuyển khoản thành công {{amount}} {{currency}} đến {{merchant}}. Mã giao dịch: {{transaction_id}}') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- WebSocket - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('PAYMENT_SUCCESS',
        'WEB_SOCKET',
        'en',
        'Payment Successful',
        'You have successfully transferred {{amount}} {{currency}} to {{merchant}}. Transaction ID: {{transaction_id}}') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Push Notification - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('PAYMENT_SUCCESS',
        'PUSH_NOTIFICATION',
        'vi',
        '💰 Thanh toán thành công',
        'Chuyển khoản {{amount}} VND đến {{merchant}} thành công!') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Push Notification - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('PAYMENT_SUCCESS',
        'PUSH_NOTIFICATION',
        'en',
        '💰 Payment Successful',
        'Successfully transferred {{amount}} {{currency}} to {{merchant}}!') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- SMS - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('PAYMENT_SUCCESS',
        'SMS',
        'vi',
        'Thong bao giao dich',
        'Ban da chuyen {{amount}} VND den {{merchant}}. Ma GD: {{transaction_id}}. Chi tiet tai app.') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- SMS - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('PAYMENT_SUCCESS',
        'SMS',
        'en',
        'Transaction notification',
        'You transferred {{amount}} {{currency}} to {{merchant}}. Transaction ID: {{transaction_id}}. Details in app.') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Email - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('PAYMENT_SUCCESS',
        'EMAIL',
        'vi',
        'Xác nhận giao dịch thành công',
        '<h2>Giao dịch thành công!</h2><p>Kính gửi quý khách,</p><p>Chúng tôi xác nhận giao dịch của bạn đã được thực hiện thành công:</p><ul><li><strong>Số tiền:</strong> {{amount}} {{currency}}</li><li><strong>Người nhận:</strong> {{merchant}}</li><li><strong>Mã giao dịch:</strong> {{transaction_id}}</li><li><strong>Số tài khoản:</strong> {{account_number}}</li></ul><p>Cảm ơn bạn đã sử dụng dịch vụ!</p>') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Email - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('PAYMENT_SUCCESS',
        'EMAIL',
        'en',
        'Transaction Confirmation',
        '<h2>Payment Successful!</h2><p>Dear Customer,</p><p>We confirm that your transaction has been completed successfully:</p><ul><li><strong>Amount:</strong> {{amount}} {{currency}}</li><li><strong>Recipient:</strong> {{merchant}}</li><li><strong>Transaction ID:</strong> {{transaction_id}}</li><li><strong>Account Number:</strong> {{account_number}}</li></ul><p>Thank you for using our service!</p>') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- OTP_CODE Templates
-- =====================================================

-- WebSocket - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('OTP_CODE',
        'WEB_SOCKET',
        'vi',
        'Mã xác thực OTP',
        'Mã OTP của bạn là: {{code}}. Loại: {{type}}. Có hiệu lực trong {{expires_in}} giây. Số lần thử còn lại: {{attempts_left}}') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- WebSocket - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('OTP_CODE',
        'WEB_SOCKET',
        'en',
        'OTP Verification Code',
        'Your OTP code is: {{code}}. Type: {{type}}. Valid for {{expires_in}} seconds. Attempts left: {{attempts_left}}') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Push Notification - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('OTP_CODE',
        'PUSH_NOTIFICATION',
        'vi',
        '🔐 Mã OTP của bạn',
        'Mã OTP: {{code}}. Hết hạn sau {{expires_in}}s') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Push Notification - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('OTP_CODE',
        'PUSH_NOTIFICATION',
        'en',
        '🔐 Your OTP Code',
        'OTP Code: {{code}}. Expires in {{expires_in}}s') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- SMS - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('OTP_CODE',
        'SMS',
        'vi',
        'Ma OTP',
        'Ma OTP cua ban la: {{code}}. Hieu luc trong {{expires_in}}s. KHONG chia se ma nay!') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- SMS - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('OTP_CODE',
        'SMS',
        'en',
        'OTP Code',
        'Your OTP code is: {{code}}. Valid for {{expires_in}}s. DO NOT share this code!') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Email - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('OTP_CODE',
        'EMAIL',
        'vi',
        'Mã xác thực OTP của bạn',
        '<h2>🔐 Mã OTP xác thực</h2><p>Mã OTP của bạn là:</p><h1 style="font-size: 32px; color: #4F46E5; letter-spacing: 8px;">{{code}}</h1><p><strong>Loại xác thực:</strong> {{type}}</p><p><strong>Thời gian hiệu lực:</strong> {{expires_in}} giây</p><p><strong>Số lần thử còn lại:</strong> {{attempts_left}}</p><p style="color: red;"><strong>⚠️ CẢNH BÁO:</strong> Không chia sẻ mã này với bất kỳ ai!</p>') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Email - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('OTP_CODE',
        'EMAIL',
        'en',
        'Your OTP Verification Code',
        '<h2>🔐 OTP Verification</h2><p>Your OTP code is:</p><h1 style="font-size: 32px; color: #4F46E5; letter-spacing: 8px;">{{code}}</h1><p><strong>Verification Type:</strong> {{type}}</p><p><strong>Valid For:</strong> {{expires_in}} seconds</p><p><strong>Attempts Left:</strong> {{attempts_left}}</p><p style="color: red;"><strong>⚠️ WARNING:</strong> Do not share this code with anyone!</p>') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- BALANCE_FLUCTUATION Templates
-- =====================================================

-- WebSocket - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('BALANCE_FLUCTUATION',
        'WEB_SOCKET',
        'vi',
        'Biến động số dư',
        'Số dư tài khoản của bạn đã thay đổi từ {{old_balance}} VND thành {{new_balance}} VND. Thay đổi: {{change}} VND') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- WebSocket - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('BALANCE_FLUCTUATION',
        'WEB_SOCKET',
        'en',
        'Balance Update',
        'Your account balance has changed from {{old_balance}} VND to {{new_balance}} VND. Change: {{change}} VND') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Push Notification - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('BALANCE_FLUCTUATION',
        'PUSH_NOTIFICATION',
        'vi',
        '💳 Cập nhật số dư',
        'Số dư mới: {{new_balance}} VND ({{change}} VND)') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Push Notification - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('BALANCE_FLUCTUATION',
        'PUSH_NOTIFICATION',
        'en',
        '💳 Balance Update',
        'New balance: {{new_balance}} VND ({{change}} VND)') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- SMS - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('BALANCE_FLUCTUATION',
        'SMS',
        'vi',
        'Cap nhat so du',
        'So du TK: {{new_balance}} VND. Thay doi: {{change}} VND. Chi tiet tai app.') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- SMS - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('BALANCE_FLUCTUATION',
        'SMS',
        'en',
        'Balance update',
        'Account balance: {{new_balance}} VND. Change: {{change}} VND. Details in app.') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Email - Vietnamese
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('BALANCE_FLUCTUATION',
        'EMAIL',
        'vi',
        'Thông báo biến động số dư tài khoản',
        '<h2>💳 Biến động số dư tài khoản</h2><p>Kính gửi quý khách,</p><p>Số dư tài khoản của bạn đã có thay đổi:</p><ul><li><strong>Số dư cũ:</strong> {{old_balance}} VND</li><li><strong>Số dư mới:</strong> {{new_balance}} VND</li><li><strong>Thay đổi:</strong> {{change}} VND</li></ul><p>Nếu bạn không thực hiện giao dịch này, vui lòng liên hệ hotline ngay!</p>') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- Email - English
INSERT INTO notification_templates (event_type, channel, language, title_template, content_template)
VALUES ('BALANCE_FLUCTUATION',
        'EMAIL',
        'en',
        'Account Balance Update Notification',
        '<h2>💳 Account Balance Update</h2><p>Dear Customer,</p><p>Your account balance has been updated:</p><ul><li><strong>Previous Balance:</strong> {{old_balance}} VND</li><li><strong>New Balance:</strong> {{new_balance}} VND</li><li><strong>Change:</strong> {{change}} VND</li></ul><p>If you did not make this transaction, please contact our hotline immediately!</p>') ON CONFLICT (event_type, channel, language) DO
UPDATE SET
    title_template = EXCLUDED.title_template,
    content_template = EXCLUDED.content_template;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Delete all templates (rollback)
DELETE
FROM notification_templates
WHERE event_type IN ('PAYMENT_SUCCESS', 'OTP_CODE', 'BALANCE_FLUCTUATION');

-- +goose StatementEnd