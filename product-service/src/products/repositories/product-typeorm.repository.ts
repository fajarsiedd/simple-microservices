import { Injectable } from "@nestjs/common";
import { InjectRepository } from "@nestjs/typeorm";
import { Repository } from "typeorm";
import { CreateProductDto } from "../dto/create-product.dto";
import { UpdateProductDto } from "../dto/update-product.dto";
import { Product } from "../entities/product.entity";
import { IProductRepository } from "../interfaces/product-repository.interface";

@Injectable()
export class ProductTypeOrmRepository implements IProductRepository {
  constructor(
    @InjectRepository(Product)
    private readonly productRepository: Repository<Product>,
  ) { }

  async create(productData: CreateProductDto): Promise<Product> {
    return this.productRepository.save(productData);
  }

  async findAll(): Promise<Product[]> {
    return this.productRepository.find();
  }

  async findOneById(id: number): Promise<Product | null> {
    return this.productRepository.findOneBy({ id });
  }

  async update(id: number, productData: UpdateProductDto): Promise<Product | null> {
    const product = await this.productRepository.findOneBy({ id });

    if (!product) {
      return null;
    }

    const updatedProduct = { ...product, ...productData };
    await this.productRepository.save(updatedProduct);

    return this.productRepository.findOneBy({ id });
  }

  async remove(id: number): Promise<void> {
    await this.productRepository.delete(id);
  }
}