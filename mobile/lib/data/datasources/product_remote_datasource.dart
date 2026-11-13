// ===== FILE: mobile/lib/data/datasources/product_remote_datasource.dart =====
import 'package:dio/dio.dart';
import '../../core/network/api_response.dart';
import '../../core/constants/app_constants.dart';
import '../models/product_model.dart';

class ProductRemoteDataSource {
  final Dio _dio;

  ProductRemoteDataSource(this._dio);

  Future<ProductModel> getProductByCode(String code) async {
    try {
      final response = await _dio.get(
        ApiEndpoints.productByCode(code),
      );

      final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
        response.data,
        (data) => data as Map<String, dynamic>,
      );

      if (apiResponse.success && apiResponse.data != null) {
        return ProductModel.fromJson(apiResponse.data!);
      } else {
        throw ApiException(
          message: apiResponse.error?.message ?? 'Product not found',
          code: apiResponse.error?.code,
        );
      }
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) {
        throw ApiException(
          message: 'Produk tidak ditemukan',
          statusCode: 404,
        );
      }
      throw ApiException(
        message: e.response?.data['error']?['message'] ?? e.message ?? 'Network error',
        statusCode: e.response?.statusCode,
      );
    }
  }

  Future<List<ProductModel>> searchProducts(String query) async {
    try {
      final response = await _dio.get(
        ApiEndpoints.products,
        queryParameters: {'search': query},
      );

      final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
        response.data,
        (data) => data as Map<String, dynamic>,
      );

      if (apiResponse.success && apiResponse.data != null) {
        final products = apiResponse.data!['products'] as List;
        return products.map((json) => ProductModel.fromJson(json)).toList();
      } else {
        throw ApiException(
          message: apiResponse.error?.message ?? 'Search failed',
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
}
// ===== END FILE =====
