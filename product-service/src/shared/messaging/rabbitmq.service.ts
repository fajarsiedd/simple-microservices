import { Inject, Injectable } from '@nestjs/common';
import { ClientProxy } from '@nestjs/microservices';

@Injectable()
export class RabbitmqService {
  constructor(@Inject('RABBITMQ_CLIENT') private readonly client: ClientProxy) {}

  public async publishEvent(topic: string, data: any) {
    this.client.emit(topic, data);
  }
}