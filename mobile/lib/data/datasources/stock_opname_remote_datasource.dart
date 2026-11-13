// ===== FILE: mobile/lib/data/datasources/stock_opname_remote_datasource.dart =====
import 'package:dio/dio.dart';
import '../../core/network/api_response.dart';
import '../../core/constants/app_constants.dart';
import '../models/stock_opname_model.dart';

class StockOpnameRemoteDataSource {
  final Dio _dio;

  StockOpnameRemoteDataSource(this._dio);

  Future<List<StockOpnameSessionModel>> getSessions() async {
    try {
      final response = await _dio.get(ApiEndpoints.stockOpnameSessions);

      final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
        response.data,
        (data) => data as Map<String, dynamic>,
      );

      if (apiResponse.success && apiResponse.data != null) {
        final sessions = apiResponse.data!['sessions'] as List;
        return sessions.map((json) => StockOpnameSessionModel.fromJson(json)).toList();
      } else {
        throw ApiException(
          message: apiResponse.error?.message ?? 'Failed to get sessions',
          code: apiResponse.error?.code,
        );
      }
    } on DioException catch (e) {
      throw ApiException(
        message: e.response?.data['error']?['message'] ?? e.message ?? 'Network error',
        statusCode: e.response?.statusCode,
      );
    }
  }

  Future<StockOpnameSessionModel> getSession(String sessionId) async {
    try {
      final response = await _dio.get(
        ApiEndpoints.stockOpnameSession(sessionId),
      );

      final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
        response.data,
        (data) => data as Map<String, dynamic>,
      );

      if (apiResponse.success && apiResponse.data != null) {
        return StockOpnameSessionModel.fromJson(apiResponse.data!);
      } else {
        throw ApiException(
          message: apiResponse.error?.message ?? 'Failed to get session',
          code: apiResponse.error?.code,
        );
      }
    } on DioException catch (e) {
      throw ApiException(
        message: e.response?.data['error']?['message'] ?? e.message ?? 'Network error',
        statusCode: e.response?.statusCode,
      );
    }
  }

  Future<List<StockOpnameDetailModel>> getSessionDetails(String sessionId) async {
    try {
      final response = await _dio.get(
        ApiEndpoints.stockOpnameSessionDetails(sessionId),
      );

      final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
        response.data,
        (data) => data as Map<String, dynamic>,
      );

      if (apiResponse.success && apiResponse.data != null) {
        final details = apiResponse.data!['details'] as List;
        return details.map((json) => StockOpnameDetailModel.fromJson(json)).toList();
      } else {
        throw ApiException(
          message: apiResponse.error?.message ?? 'Failed to get session details',
          code: apiResponse.error?.code,
        );
      }
    } on DioException catch (e) {
      throw ApiException(
        message: e.response?.data['error']?['message'] ?? e.message ?? 'Network error',
        statusCode: e.response?.statusCode,
      );
    }
  }

  Future<StockOpnameDetailModel> submitScan(
    String sessionId,
    SubmitScanRequest request,
  ) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.stockOpnameSubmitScan(sessionId),
        data: request.toJson(),
      );

      final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
        response.data,
        (data) => data as Map<String, dynamic>,
      );

      if (apiResponse.success && apiResponse.data != null) {
        return StockOpnameDetailModel.fromJson(apiResponse.data!);
      } else {
        throw ApiException(
          message: apiResponse.error?.message ?? 'Failed to submit scan',
          code: apiResponse.error?.code,
        );
      }
    } on DioException catch (e) {
      throw ApiException(
        message: e.response?.data['error']?['message'] ?? e.message ?? 'Network error',
        statusCode: e.response?.statusCode,
      );
    }
  }

  Future<void> completeSession(String sessionId) async {
    try {
      await _dio.post(ApiEndpoints.stockOpnameComplete(sessionId));
    } on DioException catch (e) {
      throw ApiException(
        message: e.response?.data['error']?['message'] ?? e.message ?? 'Network error',
        statusCode: e.response?.statusCode,
      );
    }
  }
}
// ===== END FILE =====
