import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import { TransformInterceptor } from './shared/interceptors/response.interceptor';
import { HttpExceptionFilter } from './shared/filters/http-exception.filter';
import { ValidationPipe } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Transport } from '@nestjs/microservices';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  const configService = app.get(ConfigService);
  const RABBITMQ_URL = configService.get<string>('RABBITMQ_URL', 'amqp://localhost:5672');
  const RABBITMQ_ORDER_EXCHANGE_NAME = configService.get<string>('RABBITMQ_ORDER_EXCHANGE_NAME', 'order_exchange');

  app.connectMicroservice({
    transport: Transport.RMQ,
    options: {
      urls: [RABBITMQ_URL],
      queue: 'product-service.order.created',
      queueOptions: { durable: true, autoDelete: false },
      noAck: false,
      exchange: RABBITMQ_ORDER_EXCHANGE_NAME,
      exchangeType: 'direct',
      routingKey: 'order.created',
      prefetchCount: 10,
    },
  });

  app.useGlobalPipes(new ValidationPipe({
    whitelist: true,
    forbidNonWhitelisted: true,
    transform: true,
  }));
  app.useGlobalInterceptors(new TransformInterceptor());
  app.useGlobalFilters(new HttpExceptionFilter());

  await app.startAllMicroservices();
  await app.listen(3000);
}
bootstrap();
