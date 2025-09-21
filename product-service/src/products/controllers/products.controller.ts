// src/products/products.controller.ts
import { Body, Controller, Delete, Get, Param, Post, Put } from '@nestjs/common';
import { CreateProductDto } from '../dto/create-product.dto';
import { ProductsService } from '../services/products.service';
import { Product } from '../entities/product.entity';
import { UpdateProductDto } from '../dto/update-product.dto';
import { Ctx, EventPattern, Payload, RmqContext } from '@nestjs/microservices';

@Controller('products')
export class ProductsController {
    constructor(private readonly productsService: ProductsService) { }

    @Post()
    create(@Body() createProductDto: CreateProductDto): Promise<Product> {
        return this.productsService.create(createProductDto);
    }

    @Get()
    findAll(): Promise<Product[]> {
        return this.productsService.findAll();
    }

    @Get(':id')
    findOne(@Param('id') id: number): Promise<Product | null> {
        return this.productsService.findOne(id);
    }

    @Put(':id')
    update(@Param('id') id: number, @Body() updateProductDto: UpdateProductDto): Promise<Product | null> {
        return this.productsService.update(id, updateProductDto);
    }

    @Delete(':id')
    remove(@Param('id') id: number): Promise<void> {
        return this.productsService.remove(id);
    }

    @EventPattern('order.created')
    async handleOrderCreated(
        @Payload() data: any,
        @Ctx() context: RmqContext,
    ) {
        console.log('[Event] order.created received:', data);

        try {
            await this.productsService.reduceStock(
                data.orderID,
                data.productID,
                data.qty,
            );

            const channel = context.getChannelRef();
            const originalMsg = context.getMessage();
            channel.ack(originalMsg);

        } catch (err) {
            console.error('Error processing order.created:', err);

            const channel = context.getChannelRef();
            const originalMsg = context.getMessage();
            channel.nack(originalMsg, false, true);
        }
    }
}