import { Injectable, OnModuleInit } from '@nestjs/common';
import * as amqp from 'amqplib';

@Injectable()
export class RabbitmqInitService implements OnModuleInit {
    async onModuleInit() {
        const RABBITMQ_URL = process.env.RABBITMQ_URL || 'amqp://localhost:5672';
        const ORDER_EXCHANGE = process.env.RABBITMQ_ORDER_EXCHANGE_NAME || 'order_exchange';
        const PRODUCT_EXCHANGE = process.env.RABBITMQ_PRODUCT_EXCHANGE_NAME || 'product_exchange';

        const connection = await amqp.connect(RABBITMQ_URL);
        const channel = await connection.createChannel();

        await channel.assertExchange(PRODUCT_EXCHANGE, 'direct', { durable: true });
        await channel.assertQueue('product-service.product.created', { durable: true });
        await channel.bindQueue('product-service.product.created', PRODUCT_EXCHANGE, 'product.created');

        await channel.assertExchange(ORDER_EXCHANGE, 'direct', { durable: true });
        await channel.assertQueue('product-service.order.created', { durable: true });
        await channel.bindQueue('product-service.order.created', ORDER_EXCHANGE, 'order.created');

        await channel.close();
        await connection.close();
    }
}
