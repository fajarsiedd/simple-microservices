import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { RabbitmqService } from './rabbitmq.service';
import { RabbitmqInitService } from './rabbitmq-init.service';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: 'RABBITMQ_CLIENT',
        imports: [ConfigModule],
        inject: [ConfigService],
        useFactory: async (configService: ConfigService) => ({
          transport: Transport.RMQ,
          options: {
            urls: [configService.get<string>('RABBITMQ_URL', 'amqp://localhost:5672')],
            queue: 'product-service.product.created',
            queueOptions: { durable: true }
          },
        }),
      },
    ]),
  ],
  providers: [RabbitmqService, RabbitmqInitService],
  exports: [RabbitmqService],
})
export class RabbitmqModule { }