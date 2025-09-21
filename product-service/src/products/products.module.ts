import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ProductsService } from './services/products.service';
import { ProductsController } from './controllers/products.controller';
import { Product } from './entities/product.entity';
import { ProductTypeOrmRepository } from './repositories/product-typeorm.repository';
import { IProductRepository } from './interfaces/product-repository.interface';
import { CachingModule } from '../shared/caching/caching.module';
import { RabbitmqModule } from '../shared/messaging/rabbitmq.module';

@Module({
    imports: [
        TypeOrmModule.forFeature([Product]),
        CachingModule,
        RabbitmqModule
    ],
    controllers: [ProductsController],
    providers: [
        ProductsService,
        {
            provide: 'IProductRepository',
            useClass: ProductTypeOrmRepository,
        },
    ],
})
export class ProductsModule { }