import { ExceptionFilter, Catch, ArgumentsHost, HttpException, HttpStatus } from '@nestjs/common';

@Catch(HttpException)
export class HttpExceptionFilter implements ExceptionFilter {
  catch(exception: HttpException, host: ArgumentsHost) {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse();
    const status = exception instanceof HttpException ? exception.getStatus() : HttpStatus.INTERNAL_SERVER_ERROR;
    
    const exceptionResponse = exception.getResponse();
    const error = typeof exceptionResponse === 'object' ? (exceptionResponse as any).error || exceptionResponse : exceptionResponse;
    const message = typeof exceptionResponse === 'object' ? (exceptionResponse as any).message || 'Internal Server Error' : exceptionResponse;

    response.status(status).json({
      status: false,
      message: message,
      errors: error,
      data: null,
    });
  }
}