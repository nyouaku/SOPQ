// ===== FILE: mobile/lib/core/constants/app_constants.dart =====
class AppConstants {
  // API Configuration
  // Override saat build: flutter build apk --dart-define=API_BASE_URL=http://192.168.1.10:8080/api/v1
  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080/api/v1',
  );
  static const String wsBaseUrl = String.fromEnvironment(
    'WS_BASE_URL',
    defaultValue: 'ws://localhost:8080/api/v1',
  );

  // Timeouts
  static const Duration connectionTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);

  // Storage Keys
  static const String accessTokenKey = 'access_token';
  static const String refreshTokenKey = 'refresh_token';
  static const String userIdKey = 'user_id';
  static const String userRoleKey = 'user_role';
  static const String userNameKey = 'user_name';

  // Database
  static const String dbName = 'stock_opname.db';
  static const int dbVersion = 2;

  // Offline Sync
  static const int maxRetryAttempts = 3;
  static const Duration syncInterval = Duration(minutes: 5);
  static const int batchSize = 50;

  // Scanner
  static const Duration scannerDebounce = Duration(milliseconds: 500);
  static const int maxScanHistory = 100;

  // Voice Input
  static const Duration voiceListenTimeout = Duration(seconds: 10);
  static const String voiceLocale = 'id_ID';

  // Pagination
  static const int defaultPageSize = 20;

  // Performance
  static const int searchDebounceMs = 300;
  static const int cacheMaxAge = 3600; // seconds
}

class ApiEndpoints {
  // Auth
  static const String login = '/auth/login';
  static const String logout = '/auth/logout';
  static const String refreshToken = '/auth/refresh';
  static const String profile = '/auth/me';

  // Products
  static const String products = '/products';
  static String productByCode(String code) => '/products/code/$code';

  // Stock Opname
  static const String stockOpnameSessions = '/stock-opname/sessions';
  static String stockOpnameSession(String id) => '/stock-opname/sessions/$id';
  static String stockOpnameSessionDetails(String id) => '/stock-opname/sessions/$id/details';
  static String stockOpnameSubmitScan(String id) => '/stock-opname/sessions/$id/scan';
  static String stockOpnameComplete(String id) => '/stock-opname/sessions/$id/complete';

  // Dashboard
  static const String dashboardSummary = '/dashboard/summary';

  // Notifications
  static const String notifications = '/notifications';
  static String markNotificationRead(String id) => '/notifications/$id/read';

  // WebSocket
  static String stockOpnameWS(String sessionId) => '/ws/stock-opname/sessions/$sessionId';
}
// ===== END FILE =====
