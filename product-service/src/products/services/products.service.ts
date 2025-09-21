import { Injectable, Inject, NotFoundException } from '@nestjs/common';
import { Product } from '../entities/product.entity';
import type { IProductRepository } from '../interfaces/product-repository.interface';
import { CreateProductDto } from '../dto/create-product.dto';
import { RedisService } from '@liaoliaots/nestjs-redis'
import { UpdateProductDto } from '../dto/update-product.dto';
import { RabbitmqService } from '../../shared/messaging/rabbitmq.service';
import Redis from 'ioredis';
import { InjectDataSource } from '@nestjs/typeorm';
import { DataSource } from "typeorm";

@Injectable()
export class ProductsService {
    private client: Redis;

    constructor(
        @Inject('IProductRepository')
        private readonly productRepository: IProductRepository,
        private readonly redisService: RedisService,
        private readonly rabbitmqService: RabbitmqService,
        @InjectDataSource()
        private dataSource: DataSource
    ) {
        this.client = this.redisService.getOrThrow('default');
    }

    async create(productData: CreateProductDto): Promise<Product> {
        const newProduct = await this.productRepository.create(productData);

        if (newProduct.id) {
            await this.client.del(`products:id:${newProduct.id}`);
        }

        this.rabbitmqService.publishEvent('product.created', newProduct);

        return newProduct;
    }

    async findAll(): Promise<Product[]> {
        return this.productRepository.findAll();
    }

    async findOne(id: number): Promise<Product | null> {
        const cacheKey = `products:id:${id}`;

        const cachedProduct = await this.client.get(cacheKey);
        if (cachedProduct) {
            return JSON.parse(cachedProduct);
        }

        const product = await this.productRepository.findOneById(id);

        if (product) {
            await this.client.set(cacheKey, JSON.stringify(product));

            return product;
        } else {
            throw new NotFoundException(`Product with ID ${id} not found.`);
        }
    }

    async update(id: number, productData: UpdateProductDto): Promise<Product | null> {
        const updatedProduct = await this.productRepository.update(id, productData);

        if (updatedProduct) {
            await this.client.del(`products:id:${updatedProduct.id}`);
            return updatedProduct;
        } else {
            throw new NotFoundException(`Product with ID ${id} not found.`);
        }
    }

    async remove(id: number): Promise<void> {
        await this.client.del(`products:id:${id}`);

        await this.productRepository.remove(id);
    }

    async reduceStock(orderId: number, productId: number, qty: number): Promise<void> {
        await this.dataSource.transaction(async manager => {
            const product = await manager.getRepository(Product)
                .createQueryBuilder('product')
                .setLock('pessimistic_write')
                .where('product.id = :id', { id: productId })
                .getOne();

            if (!product) {
                return;
            }

            const newQty = product.qty - qty;
            if (newQty < 0) {
                return;
            }

            product.qty = newQty;
            await manager.save(product);

            await this.client.del(`products:id:${product.id}`);
        });
    }
}